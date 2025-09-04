import os
import json
import time
import requests
import redis
from kafka import KafkaProducer

KAFKA_BROKERS = os.getenv("KAFKA_BROKERS", "kafka:9092")
OPPS_TOPIC = os.getenv("OPPORTUNITIES_TOPIC", "opportunities")
FEESFX_URL = os.getenv("FEESFX_URL", "http://fees-fx:8090/compute")
REDIS_ADDR = os.getenv("REDIS_ADDR", "redis:6379")
POLL_INTERVAL = int(os.getenv("POLL_INTERVAL", "3"))  # seconds
ITEM_KEYS = [i for i in os.getenv("ITEM_KEYS", "").split(",") if i]

redis_cli = redis.Redis.from_url(f"redis://{REDIS_ADDR}", decode_responses=True)

producer = KafkaProducer(
    bootstrap_servers=KAFKA_BROKERS.split(","),
    value_serializer=lambda v: json.dumps(v).encode("utf-8"),
)


def fetch_quotes(items):
    out = {}
    for item in items:
        val = redis_cli.get(f"quotes:{item}")
        if not val:
            continue
        try:
            data = json.loads(val)
        except Exception:
            continue

        if isinstance(data, list):
            qmap = {q["market"]: q for q in data if isinstance(q, dict) and "market" in q}
        elif isinstance(data, dict):
            # allow straight mapping market->quote
            qmap = {k: v for k, v in data.items() if isinstance(v, dict)}
        else:
            continue
        out[item] = qmap
    return out


while True:
    quotes = fetch_quotes(ITEM_KEYS)
    for item, markets_data in quotes.items():
        markets = list(markets_data.keys())
        if len(markets) < 2:
            continue

        for i in range(len(markets)):
            for j in range(i + 1, len(markets)):
                m1, m2 = markets[i], markets[j]
                p1, c1 = markets_data[m1]["price"], markets_data[m1]["currency"]
                p2, c2 = markets_data[m2]["price"], markets_data[m2]["currency"]

                try:
                    r = requests.get(
                        FEESFX_URL,
                        params={
                            "buy_price": p1,
                            "buy_currency": c1,
                            "sell_price": p2,
                            "sell_currency": c2,
                        },
                        timeout=2,
                    )
                    if r.status_code == 200:
                        data = r.json()
                        if data.get("spread_bps", 0) > 100:
                            opp = {
                                "item_key": item,
                                "buy_market": m1,
                                "sell_market": m2,
                                "spread_bps": data["spread_bps"],
                                "ts": time.time(),
                            }
                            producer.send(OPPS_TOPIC, opp)
                            print("Opportunity:", opp)
                except Exception as e:
                    print("error:", e)

    time.sleep(POLL_INTERVAL)

