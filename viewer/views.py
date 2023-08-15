from django.shortcuts import render
import boto3
import os

s3 = boto3.resource('s3', endpoint_url=os.environ.get('S3_URL'))
bufo_bucket = s3.Bucket(os.environ.get('S3_BUFO_BUCKET'))

def index(request):
    all_the_bufos = [obj.key for obj in bufo_bucket.objects.all()]
    return render(request, 'index.html', {
        'all_the_bufos': all_the_bufos,
        'bufo_url': os.environ.get('BUFO_URL')
    })
