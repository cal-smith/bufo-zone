from django.shortcuts import render
from django.http import JsonResponse, request
from django.views.decorators.csrf import ensure_csrf_cookie
import boto3
import os
import json
from .models import Bufo

s3 = boto3.resource('s3', endpoint_url=os.environ.get('S3_URL'))
bufo_bucket = s3.Bucket(os.environ.get('S3_BUFO_BUCKET'))

@ensure_csrf_cookie
def index(request):
    all_the_bufos = [
        {'name': obj.name, 'score': obj.score(), 'frogs': '🐸'*int(obj.score() or 0),'url': obj.get_url()}
        for obj in Bufo.objects.select_related().all()
    ]

    return render(request, 'index.html', {
        'all_the_bufos': all_the_bufos
    })


def vote(request: request.HttpRequest):
    if 'votes' not in request.session:
       request.session['votes'] = {}
       request.session.set_expiry(24*60*60)

    data = json.loads(request.body)
    
    if data['name'] in request.session['votes']:
        return JsonResponse({
            'error': 'already_voted',
            'description': 'Already rated this Bufo today, try again tomorrow!'
        })

    bufo = Bufo.objects.get(name=data["name"])
    bufo.bufo_vote_set.create(value=int(data["value"]))
    
    request.session['votes'][data['name']] = True
    request.session.modified = True

    return JsonResponse({
        'name': bufo.name,
        'score': bufo.score()
    })
