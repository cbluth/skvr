# Simple Key-Value Rest Server
SKVR: a key-value persistent store with a simple rest interface.

---

## What is SKVR? (say like "skivver")
SKVR builds on [boltdb](https://github.com/etcd-io/bbolt) to provide a key-value database with a simple rest interface.
You can use any standard http client (wget, curl, browser) to interact with skvr.

It supports multiple buckets or namespaces, which can contain an arbitrary amount of keys.
If you upload html/css/javascript data into specified paths/namespaces/keys, and load the url in your browser, it acts like a web server. If there is no key in the request, then the value of the index key is returned, by default it is "index.html".

## How to install?
Just install the "normal" go-esque way
```bash
go get github.com/cbluth/skvr
```

## How to run it?
Just run it from the cli, then as a client use curl, postman, or any other rest client (wget, httpie, etc)
```bash
# after building/installing
$ export SKVR_PORT=8077
$ skvr
2021/01/17 16:04:44 Simple KV Rest Server starting on port 8077...
[...]
```

## How does it work?
SKVR supports these CRUD-esque http methods:
- GET
- POST/PUT (these do the same thing)
- DELETE
- OPTIONS

The basic operations of skvr are:
- Read Key-Value (http GET)
- Write Key-Value (http POST/PUT) (will overwrite existing value)
- Delete Namespace and/or Key-Value(s) (http DELETE)
- List Namespaces or Keys (http GET/OPTIONS)

Optionally, you can configure skvr with Environment Variables:
```bash
$ export SKVR_DIR="/some/path"          # (default value is "/var/lib/skvr")
$ export SKVR_INDEX_KEY=".index"        # (default value is "index.html")
$ export SKVR_DEFAULT_NAMESPACE="myapp" # (default value is "default")
$ export SKVR_PORT="9000"               # (default value is "8077")
```

The key and namespace is inferred from the path on the request. If no namespace is given, then the default namespace is inferred. If no key is given, then the response changes depending on if the index key is existing in the namespace or not. If no key is given and the index key contains data, then the data is returned. If the index key is missing, then the response contains a list of keys available in the namespace.

The SKVR URL:
- `http://skvrhost:8077/`                  # namespace nor key is given, if the index key in the default namespace contains data, then return that data, otherwise list the namespaces 
- `http://skvrhost:8077/<key>`             # no namespace is given, then it assumes the default namespace
- `http://skvrhost:8077/<namespace>/`      # no key is given, inferred request is to list the available keys if the index key is missing, otherwise return data of index key
- `http://skvrhost:8077/<namespace>/<key>` # this request is clear
- `http://skvrhost:8077/<namespace>/<key>/<any/random-extra/paths>` # this namespace and key are still the first two in the path, and the same data is returned

Write a Value (POST/PUT)
---
```bash
# writing a value to skvr is easy, just try it with curl using PUT or POST:
curl -XPUT -d "$(date)" http://skvrhost:8077/mydate

# write some data to your key (assumes default namespace), these two PUT requests are effectively identical
curl -XPUT -d "some data" http://skvrhost:8077/your-key # effectively identical
curl -XPUT -d "some data" http://skvrhost:8077/default/your-key # effectively identical
curl -XGET http://skvrhost:8077/your-key # some data
curl -XGET http://skvrhost:8077/default/your-key # same data

# write some data to your key in your namespace, namespaces are created automatically if not existing
curl -XPUT -d "some data" http://skvrhost:8077/YourNamespace/YourKey # key and namespace are created
curl -XGET http://skvrhost:8077/YourNamespace/YourKey # some data

# use stdin with binary-data
cat some-file.zip | curl -XPUT --data-binary @- http://skvrhost:8077/backup/myfile.zip
curl -XGET http://skvrhost:8077/backup/myfile.zip | md5sum
md5sum some-file.zip # match

# doing a put/post request without a key is illegal.
curl -XPUT -d "some data" http://skvrhost:8077/ # 405 not allowed
curl -XPUT -d "some data" http://skvrhost:8077/some-namespace/ # 405 not allowed
```

Read a Value (GET)
---
```bash
# list namespaces being used (only if index key in default namespace is missing)
curl -XGET http://skvrhost:8077/ # lists namespaces
curl -XPUT -d ${RANDOM} http://skvrhost:8077/${RANDOM}/${RANDOM} # add some random namespaces
curl -XGET http://skvrhost:8077/ # lists namespaces, there are more

# overwrite index key in default namespace, and then the list functionality is disabled, using https://github.com/Y2Z/monolith
monolith https://www.bbc.com/news | curl -XPUT --data-binary @- http://skvrhost:8077/default/index.html
curl -XGET http://skvrhost:8077/ # this is better to open in a real browser,
xdg-open http://skvrhost:8077/ # list functionality is disabled, and the request returns the html page

# get value of your key in default namespace, these requests are effectively identical
curl -XGET http://skvrhost:8077/YourKey # some data
curl -XGET http://skvrhost:8077/default/YourKey # same data
curl -XGET http://skvrhost:8077/default/YourKey/any-random-path # same data, namespace and key needs to stay constant

# list keys in specific namespace (only if index key is missing)
curl -XGET http://skvrhost:8077/some-existing-namespace/ # shows keys

# get value of your key in your namespace
curl -XGET http://skvrhost:8077/some-existing-namespace/YourKey # some data

# 404 not found if namespace or key is missing
curl -XGET http://skvrhost:8077/non-existing-key # 404 error
curl -XGET http://skvrhost:8077/some-namespace/non-existing-key # 404 error
curl -XGET http://skvrhost:8077/non-existing-namespace/ # 404 error
```

Delete a Key or Namespace (DELETE)
---
```bash
# writing a value, then delete it, using the default namespace
curl -XPUT -d "$(date)" http://skvrhost:8077/mydate
curl -XDELETE http://skvrhost:8077/mydate

# write a value in a specific namespace, then delete it
curl -XPUT -d "$(date)" http://skvrhost:8077/some-namespace/mydate
curl -XDELETE http://skvrhost:8077/some-namespace/mydate

# write multiple keys in a namespace, delete the whole namespace at once
curl -XPUT -d ${RANDOM} http://skvrhost:8077/some-namespace/${RANDOM}
curl -XPUT -d ${RANDOM} http://skvrhost:8077/some-namespace/${RANDOM}
curl -XDELETE http://skvrhost:8077/some-namespace/ # the trailing slash `/` is important

# you cannot delete multiple namespaces
curl -XDELETE http://skvrhost:8077/ # 405 not allowed
```

Listing Keys, Namespaces, and Checking if they exist (OPTIONS)
---
```bash
# list the namespaces in skvr
curl --fail -XOPTIONS http://skvrhost:8077/

# check if key in default namespace exists
curl --fail -XOPTIONS http://skvrhost:8077/existing-key # success
curl --fail -XOPTIONS http://skvrhost:8077/default/existing-key # success
curl --fail -XOPTIONS http://skvrhost:8077/non-existing-key # 404 fail

# list keys in namespace
curl --fail -XOPTIONS http://skvrhost:8077/some-namespace/ # success
curl --fail -XOPTIONS http://skvrhost:8077/non-existing-namespace/ # 404 fail

# check if key exists
curl --fail -XOPTIONS http://skvrhost:8077/some-namespace/existing-key # success
curl --fail -XOPTIONS http://skvrhost:8077/some-namespace/non-existing-key # 404 fail
```

---
# TODO
- add https support
- add basic-auth support
- include cli client (eg: go get github.com/cbluth/skvr/cmd/skvr)
- include golang client module
- create/embed web ui
- tests
- benchmarks
- 

