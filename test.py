import garlic_pear
import tempfile
from pathlib import Path
from garlic_pear.utils import make_embedded_file
import os


from garlic_pear.collections.directory import DirectoryCollection



with tempfile.TemporaryDirectory() as tmpdirname:
    
    tmpdir = Path(tmpdirname)

    files = (( "a.txt", "first")
    , ( "b.txt", "second")
    , ( "c/b.txt", "third"))
    for path, data in files:
        with make_embedded_file(tmpdir / "a" / path, "w") as fd:
            fd.write(data)
        with make_embedded_file(tmpdir / "b" / path, "w") as fd:
            fd.write(data)


    os.system("find {}".format(tmpdirname))



    td = garlic_pear.TrackedDirectory(tmpdir / "a")
    ds = DirectorySource(tmpdir / "b")
    
    
    print(td.all_entries())
    print(td.entries().sort())
    for entry in td.all_entries_iter():
        print(str(entry), td.relative_path(entry))
    print(ds.entries())
    for entry in ds.all_entries_iter():
        print(str(entry))
