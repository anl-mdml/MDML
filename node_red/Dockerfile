FROM nodered/node-red
USER root
WORKDIR /usr/src/node-red/
RUN apk update && apk add wget && rm -rf /var/cache/apk/*
RUN apk add py-setuptools
RUN apk add python3-dev
RUN apk add openssl-dev
RUN apk add libffi-dev
RUN apk add imagemagick
RUN apk add graphicsmagick
RUN apk add build-base libzmq musl-dev zeromq-dev
RUN python3 -m pip install --upgrade pip
RUN pip install --user funcx
RUN pip install --user numpy
RUN pip install --user pandas==1.0.5
RUN wget https://sourceforge.net/projects/s3tools/files/s3cmd/2.0.1/s3cmd-2.0.1.tar.gz
RUN tar xzf s3cmd-2.0.1.tar.gz
WORKDIR /usr/src/node-red/s3cmd-2.0.1/
RUN python3 setup.py install
RUN ln -sf /usr/bin/python3 /usr/bin/python
USER node-red
RUN npm install --save gm
RUN npm install --save minio
RUN npm install --save influx
RUN npm install --save base-64
RUN npm install --save melt-data
RUN npm install --save dateformat
RUN npm install --save node-red-contrib-https
RUN npm install --save node-red-contrib-batcher
RUN npm install --save node-red-contrib-influxdb
RUN npm install --save node-red-contrib-re-postgres
RUN npm install --save node-red-contrib-multipart-stream-encoder
RUN npm install --save node-red-contrib-python3-function
CMD ["npm", "start", "--", "--userDir", "/data"]
