import pathlib
import itertools


def make_embedded_file(path: pathlib.Path, *args, **kwargs):
    if not isinstance(path,pathlib.Path):
        path = pathlib.Path(path)

    parent_path = path.parent
    if not parent_path.exists():
        parent_path.mkdir(parents=True, exist_ok=True)

    return path.open( *args, **kwargs)
    








