from django.core.management.base import BaseCommand
import boto3
import os
from dotenv import load_dotenv, find_dotenv
from viewer.models import Bufo

env_file = find_dotenv()
if env_file:
    load_dotenv(env_file)

client = boto3.client('s3', endpoint_url=os.environ.get('S3_URL'))

class Command(BaseCommand):
    def handle(self, *args, **kwargs):
        paginator = client.get_paginator('list_objects')
        page_iterator = paginator.paginate(Bucket=os.environ.get('S3_BUFO_BUCKET'), PaginationConfig={'PageSize': 100})
        for page in page_iterator:
            all_the_bufos = [Bufo(name=item['Key']) for item in page['Contents']]
            Bufo.objects.bulk_create(all_the_bufos, ignore_conflicts=True)
