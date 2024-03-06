import datetime
import logging
import os
import sys
import time
import subprocess


def get_number_of_flows():
    try:
        output = subprocess.run(
            ["ovs-ofctl", "dump-aggregate", "br-int"], capture_output=True, text=True
        )
        result = output.stdout
        return int(result.split("flow_count=")[1])
    except Exception as e:
        logging.info(f"Failed getting flows count: {e}")
        return 0


# poll_interval in seconds, float
# convergence_period in seconds, for how long number of flows shouldn't change to consider it stable
# convergence_timeout in seconds, for how long number to wait for stabilisation before timing out
def wait_for_flows_to_stabilize(
    poll_interval, convergence_period, convergence_timeout, node_name
):
    timed_out = False
    timeout = convergence_timeout + convergence_period
    start = time.time()
    last_changed = time.time()
    flows_num = get_number_of_flows()
    while (
        time.time() - last_changed < convergence_period
        and time.time() - start < timeout
    ):
        new_flows_num = get_number_of_flows()
        if new_flows_num != flows_num:
            flows_num = new_flows_num
            last_changed = time.time()
            logging.info(f"{node_name}: {new_flows_num}")

        time.sleep(poll_interval)
    if time.time() - start >= timeout:
        timed_out = True
        logging.info(f"TIMEOUT: {node_name} {timeout} seconds passed")
    return last_changed, flows_num, timed_out


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
    for file in [
        "/var/log/openvswitch/ovn-controller.log",
        "/var/log/openvswitch/ovs-vswitchd.log",
        "/var/log/openvswitch/ovn-northd.log",
    ]:
        output = subprocess.run(["cat", file], capture_output=True, text=True)
        if len(output.stderr) != 0:
            continue
        else:
            output_lines = output.stdout.splitlines()
            for log_line in output_lines:
                if "no response to inactivity probe" in log_line:
                    concerning_logs.append(log_line)
    return concerning_logs


def main():
    node_name = os.getenv("MY_NODE_NAME")
    convergence_period = int(os.getenv("CONVERGENCE_PERIOD"))
    convergence_timeout = int(os.getenv("CONVERGENCE_TIMEOUT"))

    logging.basicConfig(
        format="%(asctime)s %(levelname)-8s %(message)s",
        level=logging.INFO,
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    logging.info(
        f"Start openflow-tracker {node_name}, convergence_period {convergence_period}, convergence timeout {convergence_timeout}"
    )
    stabilize_time, flow_num, timed_out = wait_for_flows_to_stabilize(
        1, convergence_period, convergence_timeout, node_name
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
    sys.exit(int(timed_out))


if __name__ == "__main__":
    main()
