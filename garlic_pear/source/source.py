from typing import List
from ..path import Path
from ..directory import Directory

class Source(Directory):

    def last_modified(self, entry):
        raise NotImplementedError()


