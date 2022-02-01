from enum import Enum
import pathlib


class DiffEnum(Enum):
    LeftExists = 1
    RightExists = 2
    NeitherExists = 3




def diff(left: pathlib.Path, right: pathlib.Path):
    return

