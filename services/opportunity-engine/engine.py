import os, json, time
from kafka import KafkaConsumer, KafkaProducer
import requests

KAFKA_BROKERS = os.getenv("KAFKA_BROKERS", "kafka:9092")
PRICES_TOPIC = os.getenv("PRICES_TOPIC", "prices.raw")
OPPS_TOPIC = os.getenv("OPPORTUNITIES_TOPIC", "opportunities")
FEESFX_URL = os.getenv("FEESFX_URL", "http://fees-fx:8090/compute")

consumer = KafkaConsumer(
    PRICES_TOPIC,
    bootstrap_servers=KAFKA_BROKERS.split(","),
    value_deserializer=lambda m: json.loads(m.decode("utf-8")),
    group_id="opp-engine"
)

producer = KafkaProducer(
    bootstrap_servers=KAFKA_BROKERS.split(","),
    value_serializer=lambda v: json.dumps(v).encode("utf-8")
)

# простейший кэш цен
latest = {}

for msg in consumer:
    ev = msg.value
    item = ev["item_key"]
    market = ev["market"]
    price = ev["price"]
    curr = ev["currency"]
    latest.setdefault(item, {})[market] = (price, curr)

    # для каждой пары рынков считаем спред
    markets = list(latest[item].keys())
    if len(markets) < 2:
        continue

    for i in range(len(markets)):
        for j in range(i+1, len(markets)):
            m1, m2 = markets[i], markets[j]
            p1, c1 = latest[item][m1]
            p2, c2 = latest[item][m2]

            try:
                r = requests.get(FEESFX_URL, params={
                    "buy_price": p1, "buy_currency": c1,
                    "sell_price": p2, "sell_currency": c2
                }, timeout=2)
                if r.status_code == 200:
                    data = r.json()
                    if data["spread_bps"] > 100:  # >1% спред
                        opp = {
                            "item_key": item,
                            "buy_market": m1,
                            "sell_market": m2,
                            "spread_bps": data["spread_bps"],
                            "ts": time.time()
                        }
                        producer.send(OPPS_TOPIC, opp)
                        print("Opportunity:", opp)
            except Exception as e:
                print("error:", e)
