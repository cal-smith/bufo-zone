from typing import Union
from django.db import models
from django.db.models import Manager, Avg
import os
    
class Bufo(models.Model):
    name = models.TextField(primary_key=True)
    created = models.DateTimeField(auto_now_add=True)
    bufo_vote_set: Manager["BufoVote"]
    vote_avg: float
    """ set if called with get_with_scores """
    
    @classmethod
    def get_with_scores(cls):
        return cls.objects.annotate(vote_avg=Avg('bufo_vote_set__value'))
    
    def get_url(self):
        return f"{os.environ.get('BUFO_URL')}/{self.name}"

    def vote(self, value: int):
        return self.bufo_vote_set.create(value=value)

    def score(self) -> Union[float, None]:
        """ don't use when bulk querying, causes n+1 query problems """
        avg = self.bufo_vote_set.aggregate(Avg('value'))['value__avg']
        if not avg:
            return None
        return round(avg, 1)


class BufoVote(models.Model):
    bufo = models.ForeignKey(Bufo, on_delete=models.CASCADE, related_name="bufo_vote_set")
    value = models.IntegerField()
    created = models.DateTimeField(auto_now_add=True)
