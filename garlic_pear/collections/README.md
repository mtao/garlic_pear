## Intro

Each folder contains the implementation for a particular type of collection.
Each collection contains a family of entries, which can in turn be directories themselves.

An entry has two functions:
* ```path``` The path to the entry, which is typically a canonical name
* ```last_modified``` A time that can be used to compare whether an entyr is newer or older than another entry


### Collection
A collection is an entry that has contains entries of its own.
Its key method is the iterable ```entries()``` and ```entries_deep```. Helpers of ```all_entries``` and ```all_entries_deep``` are available to get every entry in dictionary form.
In many cases directories are equivalent to collections.
