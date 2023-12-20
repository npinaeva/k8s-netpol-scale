FROM python:slim
RUN apt update && \
      apt install -y curl openvswitch-switch ovn-central
COPY openflow-tracker.py openflow-tracker.py
COPY ./requirements.txt requirements.txt
RUN pip install -r requirements.txt
