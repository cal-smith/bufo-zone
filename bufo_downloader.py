import requests
from zipfile import ZipFile
import boto3
import pathlib
import os
from dotenv import load_dotenv, find_dotenv

env_file = find_dotenv()
if env_file:
    load_dotenv(env_file)

REPO_URL = 'https://github.com/knobiknows/all-the-bufo/archive/refs/heads/main.zip'

r = requests.get(REPO_URL)
s3 = boto3.resource('s3', endpoint_url=os.environ.get('S3_URL'))

with open('all_the_bufo.zip', 'wb') as bf:
    for chunk in r.iter_content(chunk_size=128):
        bf.write(chunk)

with ZipFile('all_the_bufo.zip') as many_bufo:
    bufo_bucket = s3.Bucket(os.environ.get('S3_BUFO_BUCKET'))
    existing_bufos = set([obj.key for obj in bufo_bucket.objects.all()])
    for compressed_bufo in many_bufo.filelist:
        bufo = many_bufo.extract(compressed_bufo)
        print(f'extracted bufo to {bufo}')
        if '/all-the-bufo/' in bufo:
            name = pathlib.Path(bufo).name
            if name not in existing_bufos:
                print(f'uploading bufo {name}')
                bufo_bucket.upload_file(bufo, name)
            else:
                print(f'skipping existing bufo {name}')
        else:
            print(f'skipping {bufo} since it is not a valid bufo')
