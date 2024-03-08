FROM python:slim
RUN apt update && \
      apt install -y curl iptables ipset
COPY policy-tracker.py policy-tracker.py
COPY ./requirements.txt requirements.txt
RUN pip install -r requirements.txt
