import datetime
import logging
import os
import ssl
import time
import subprocess

from opensearchpy import OpenSearch


def index_result(payload, retry_count=3):
    print(f"Indexing documents in {es_index}")
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


def get_number_of_ovs_flows():
    try:
        output = subprocess.run(
            ["ovs-ofctl", "dump-aggregate", "br-int"], capture_output=True, text=True
        )
        result = output.stdout
        return int(result.split("flow_count=")[1])
    except Exception as e:
        logging.info(f"Failed getting flows count: {e}")
        return 0


def get_number_of_logical_flows():
    output = subprocess.run(
        ["ovn-sbctl", "--no-leader-only", "--columns=_uuid", "list", "logical_flow"],
        capture_output=True,
        text=True,
    )
    if len(output.stderr) != 0:
        return 0
    output_lines = output.stdout.splitlines()
    return len(output_lines) // 2 + 1


# poll_interval in seconds, float
# stable_threshold in seconds, for how long number of flows shouldn't change to consider it stable
# timout in seconds
def wait_for_flows_to_stabilize(poll_interval, stable_threshold, timeout, node_name):
    start = time.time()
    last_changed = time.time()
    ovs_flows_num = get_number_of_ovs_flows()
    logical_flows_num = get_number_of_logical_flows()
    while (
        time.time() - last_changed < stable_threshold and time.time() - start < timeout
    ):
        new_logical_flows_num = get_number_of_logical_flows()
        if new_logical_flows_num != logical_flows_num:
            logical_flows_num = new_logical_flows_num
            last_changed = time.time()
            logging.info(f"{node_name}: logical flows={new_logical_flows_num}")
        else:
            new_ovs_flows_num = get_number_of_ovs_flows()
            if new_ovs_flows_num != ovs_flows_num:
                ovs_flows_num = new_ovs_flows_num
                last_changed = time.time()
                logging.info(f"{node_name}: OVS flows={new_ovs_flows_num}")

        time.sleep(poll_interval)
    if time.time() - start >= timeout:
        logging.info(f"TIMEOUT: {node_name} {timeout} seconds passed")
    return last_changed, ovs_flows_num


def get_db_data():
    results = {}
    for table in ["acl", "port_group", "address_set"]:
        output = subprocess.run(
            ["ovn-nbctl", "--no-leader-only", "--columns=_uuid", "list", table],
            capture_output=True,
            text=True,
        )
        if len(output.stderr) != 0:
            continue
        output_lines = output.stdout.splitlines()
        results[table] = len(output_lines) // 2 + 1
    for table in ["logical_flow"]:
        output = subprocess.run(
            ["ovn-sbctl", "--no-leader-only", "--columns=_uuid", "list", table],
            capture_output=True,
            text=True,
        )
        if len(output.stderr) != 0:
            continue
        output_lines = output.stdout.splitlines()
        results[table] = len(output_lines) // 2 + 1
    return results


def check_ovn_health():
    concerning_logs = []
    files = {"vswitchd": "/var/log/openvswitch/ovs-vswitchd.log"}
    output = subprocess.run(["ls", "/var/log/pods"], capture_output=True, text=True)
    for output_line in output.stdout.splitlines():
        if "ovnkube-node" in output_line:
            files[
                "ovn-controller"
            ] = f"/var/log/pods/{output_line}/ovn-controller/0.log"
            files["northd"] = f"/var/log/pods/{output_line}/northd/0.log"
    for name, file in files.items():
        output = subprocess.run(["cat", file], capture_output=True, text=True)
        if len(output.stderr) != 0:
            concerning_logs.append(f"failed to open {file}: {output.stderr}")
        else:
            output_lines = output.stdout.splitlines()
            for log_line in output_lines:
                if "no response to inactivity probe" in log_line:
                    s = log_line.split("stderr F ")
                    if len(s) > 1:
                        timestamp = s[1]
                    else:
                        timestamp = s[0]
                    timestamp = timestamp.split("|")[0]
                    format_string = "%Y-%m-%dT%H:%M:%S.%fZ"
                    datetime_object = datetime.datetime.strptime(
                        timestamp, format_string
                    )
                    if start_time < datetime_object:
                        concerning_logs.append(name + ": " + log_line)
    return concerning_logs


def main():
    global es_server, es_index, start_time
    es_server = os.getenv("ES_SERVER")
    es_index = os.getenv("ES_INDEX_NETPOL")
    node_name = os.getenv("MY_NODE_NAME")
    uuid = os.getenv("UUID")
    threshold = int(os.getenv("THRESHOLD"))
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
        "threshold": threshold,
        "test_metadata": os.getenv("METADATA"),
    }
    index_result(doc)

    logging.info(f"Start openflow-tracker {node_name}, threshold {threshold}")
    stabilize_time, flow_num = wait_for_flows_to_stabilize(
        1, threshold, 3600, node_name
    )
    stabilize_datetime = datetime.datetime.fromtimestamp(stabilize_time)
    nbdb_data = get_db_data()
    logging.info(
        f"RESULT: time={stabilize_datetime.isoformat(sep=' ', timespec='milliseconds')} {node_name} "
        f"finished with {flow_num} flows, nbdb data: {nbdb_data}"
    )
    ovn_health_logs = check_ovn_health()
    if len(ovn_health_logs) == 0:
        logging.info(f"HEALTHCHECK: {node_name} has no problems")
    else:
        logging.info(f"HEALTHCHECK: {node_name} has concerning logs: {ovn_health_logs}")

    doc = {
        "metricName": "convergence_tracker",
        "timestamp": datetime.datetime.now(datetime.UTC),
        "workload": "network-policy-perf",
        "uuid": uuid,
        "source_name": node_name,
        "convergence_timestamp": stabilize_datetime,
        "nbdb": nbdb_data,
        "ovs_flows": flow_num,
        "unhealthy_logs": ovn_health_logs,
    }
    index_result(doc)


if __name__ == "__main__":
    main()
