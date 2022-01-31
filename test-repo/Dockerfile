FROM python:3.8

RUN pip install papermill[all]

WORKDIR /home/app

COPY input.ipynb /home/app/input.ipynb


ENTRYPOINT ["/bin/bash", "-c"]