import sys
import json

passwd = sys.argv[1]

config = {
    "version": "9",
    "hosts": {
        "myminio": {
            "url": "http://minio:9000",
            "accessKey": "admin",
            "secretKey": passwd,
            "api": "S3v4",
            "lookup": "dns" 
        }
    }
}

with open("./mdml_register/.mc/config.json", "w") as f:
    f.write(json.dumps(config))