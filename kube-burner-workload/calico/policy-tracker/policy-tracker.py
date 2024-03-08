import datetime
import logging
import os
import ssl
import sys
import time
import subprocess

from opensearchpy import OpenSearch


def index_result(payload, retry_count=30):
    logging.info(
        f"Sending metric to es server {es_server} with index {es_index}\n{payload}"
    )
    while retry_count > 0:
        try:
            ssl_ctx = ssl.create_default_context()
            ssl_ctx.check_hostname = False
            ssl_ctx.verify_mode = ssl.CERT_NONE
            es = OpenSearch([es_server])
            es.index(index=es_index, body=payload)
            retry_count = 0
        except Exception as e:
            logging.info("Failed Indexing", e)
            logging.info("Retrying to index...")
            retry_count -= 1


def get_number_of_filter_rules():
    result = get_iptables_rules("filter")
    return result.count("\n")


def get_number_of_raw_rules():
    result = get_iptables_rules("raw")
    return result.count("\n")


def get_iptables_rules(table="filter"):
    try:
        output = subprocess.run(
            ["iptables-legacy", "--list-rules", "-t", table],
            capture_output=True,
            text=True,
        )
        return output.stdout
    except Exception as e:
        logging.error(f"Failed getting iptables rules in table {table}: {e}")
        return ""


def get_ipsets_len():
    result = get_all_ipsets()
    return result.count("\n")


def get_all_ipsets():
    try:
        output = subprocess.run(
            ["ipset", "list"],
            capture_output=True,
            text=True,
        )
        return output.stdout
    except Exception as e:
        logging.error(f"Failed listing ipsets: {e}")
        return ""


# poll_interval in seconds, float
# convergence_period in seconds, for how long number of flows shouldn't change to consider it stable
# convergence_timeout in seconds, for how long number to wait for stabilisation before timing out
def wait_for_rules_to_stabilize(
    poll_interval, convergence_period, convergence_timeout, node_name
):
    timeout = convergence_timeout + convergence_period
    start = time.time()
    last_changed = time.time()
    filter_rules_num = get_number_of_filter_rules()
    raw_rules_num = get_number_of_raw_rules()
    changed = False
    ipsets_len = get_ipsets_len()
    while time.time() - last_changed < convergence_period:
        if time.time() - start >= timeout:
            logging.info(f"TIMEOUT: {node_name} {timeout} seconds passed")
            return 1

        new_raw_rules_num = get_number_of_raw_rules()
        if new_raw_rules_num != raw_rules_num:
            raw_rules_num = new_raw_rules_num
            last_changed = time.time()
            changed = True
            logging.info(f"{node_name}: iptables raw table rules={raw_rules_num}")

        new_filter_rules_num = get_number_of_filter_rules()
        if new_filter_rules_num != filter_rules_num:
            filter_rules_num = new_filter_rules_num
            last_changed = time.time()
            changed = True
            logging.info(f"{node_name}: iptables filter table rules={filter_rules_num}")

        new_ipsets_len = get_ipsets_len()
        if new_ipsets_len != ipsets_len:
            ipsets_len = new_ipsets_len
            last_changed = time.time()
            changed = True
            logging.info(f"{node_name}: length of ipset list={ipsets_len}")

        if changed:
            doc = {
                "metricName": "convergence_tracker",
                "timestamp": datetime.datetime.now(datetime.UTC),
                "workload": "network-policy-perf",
                "uuid": uuid,
                "source_name": node_name,
                "convergence_timestamp": datetime.datetime.fromtimestamp(last_changed),
                "iptables_filter_rules": filter_rules_num,
                "iptables_raw_rules": raw_rules_num,
                "ipsets_list_len": ipsets_len,
            }
            index_result(doc)
            changed = False

        time.sleep(poll_interval)

    stabilize_datetime = datetime.datetime.fromtimestamp(last_changed)
    logging.info(
        f"RESULT: time={stabilize_datetime.isoformat(sep=' ', timespec='milliseconds')} {node_name} "
        f"finished with {filter_rules_num} rules in filter table, and {raw_rules_num} rules in raw table "
        f"and with {ipsets_len} lines in ipset list."
    )
    return 0


def main():
    global es_server, es_index, start_time, uuid
    es_server = os.getenv("ES_SERVER")
    es_index = os.getenv("ES_INDEX_NETPOL")
    node_name = os.getenv("MY_NODE_NAME")
    uuid = os.getenv("UUID")
    convergence_period = int(os.getenv("CONVERGENCE_PERIOD"))
    convergence_timeout = int(os.getenv("CONVERGENCE_TIMEOUT"))
    start_time = datetime.datetime.now()

    logging.basicConfig(
        format="%(asctime)s %(levelname)-8s %(message)s",
        level=logging.INFO,
        datefmt="%Y-%m-%d %H:%M:%S",
    )
    doc = {
        "metricName": "convergence_tracker_info",
        "timestamp": datetime.datetime.now(datetime.UTC),
        "workload": "network-policy-perf",
        "uuid": uuid,
        "source_name": node_name,
        "convergence_period": convergence_period,
        "convergence_timeout": convergence_timeout,
        "test_metadata": os.getenv("METADATA"),
    }
    index_result(doc)

    logging.info(
        f"Start calico-tracker {node_name}, convergence_period {convergence_period}, convergence timeout {convergence_timeout}"
    )
    timeout = wait_for_rules_to_stabilize(
        10, convergence_period, convergence_timeout, node_name
    )
    sys.exit(timeout)


if __name__ == "__main__":
    main()
