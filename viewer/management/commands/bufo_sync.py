from django.core.management.base import BaseCommand
import boto3
import os
from dotenv import load_dotenv, find_dotenv
from viewer.models import Bufo

env_file = find_dotenv()
if env_file:
    load_dotenv(env_file)

s3 = boto3.resource('s3', endpoint_url=os.environ.get('S3_URL'))
bufo_bucket = s3.Bucket(os.environ.get('S3_BUFO_BUCKET'))

class Command(BaseCommand):
    def handle(self, *args, **kwargs):
        all_the_bufos = [Bufo(name=obj.key) for obj in bufo_bucket.objects.all()]
        Bufo.objects.bulk_create(all_the_bufos, batch_size=100, ignore_conflicts=True)
