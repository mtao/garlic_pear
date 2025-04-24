
class Directory(FilesystemEntry):
    def __init__(self, *args, **kwargs):
        super(self,FilesystemEntry).__init__(*args,**kwargs)

    def canonical_path(self) -> pathlib.Path:
        return self.__canonical_path

    def last_modified(self):
        if self.path().exists():
            return None
        else:
            return self.path().stat().st_mtime

    def is_directory(self):
        return True 

    def is_empty(self) -> bool:
        return len(self.path().iterdir()) == 0


    # expects a relative path below this path
    def get_child_entry(self, child_path):
        cp = self.path() / child_path
        if not cp.exists():
            raise FileNotFoundError(cp)
        elif cp.is_dir():
            return FilesystemDirectory(cp)
        elif cp.is_file():
            return FilesystemEntry(cp)
        else:
            return FilesystemEntry(cp)

    # all file/directories in the directory
    def entries(self):
        for p in self.path().iterdir():
            yield self.get_child_entry(self.relative_path(p))
    

    # all leaf nodes in the directory
    def leaf_child_entries(self):
        for p in self.path().iterdir():
            e = self.get_child_entry(self.relative_path(p))
            if e.is_directory() and e.is_empty():
                for e in e.child_entries():
                    yield e
            else:
                yield e
