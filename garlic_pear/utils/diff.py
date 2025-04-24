from enum import Enum
import pathlib
from typing import Tuple,Map
from ..path import Path


class DiffEnum(Enum):
    Identical = 1
    LeftUpdated = 2
    RightUpdated = 3


class UpdateType(Enum):
    Created = 1
    Deleted = 2
    Modified = 3
    Unchanged = 4


class Update:
    def __init__(self, update_type, summary = None):
        self.type = update_type
        self.summary = summary





class DiffReport:
    def __init__(self, updates: Map[Path,Tuple[DiffEnum,Update]]):
        self.left = dict()
        self.right = dict()
        self.unchanged = set()
        for path, diff_result in updates:
            diff,update = diff_result
            if diff == DiffEnum.Identical:
                self.unchanged.add(path)
            elif diff == DiffEnum.RightUpdated:
                self.right[path] = update
            elif diff == DiffEnum.LeftUpdated:
                self.left[path] = update
            else:
                raise Exception("Inavlid enum value")

    def summary(self, view_unchanged = False):
        lines = []

        if view_unchanged:
            for path in self.unchanged:
                lines.append("= {}".format(path))
        for path in self.left.keys():
            lines.append("< {}".format(path))
        for path in self.right.keys():
            lines.append("> {}".format(path))

        return "\n".join(lines)






