<p align="center">
<a href="https://hits.seeyoufarm.com"/><img src="https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fnicewook%2Ftest-multipart"/></a>
</p>

---

# Sample codes for reqeust / response with files and data (key/value)

Sending files and data with multipart/form-data type between HTTP Server and Client

1. Run server in server directory

   ```
   bazel run //server:server
   ```

2. Run client in client directory
    ```$go run client.go```

3. Then, you can see server and client directory's files are sent to each other

4. run reset.bat to remove transferred files

---

Reference links

- server side simple parsing of `request multipart/form-data`
  - https://github.com/golang-samples/http/blob/master/fileupload/main.go
- response with `multipart/form-data`
  - https://peter.bourgon.org/blog/2019/02/12/multipart-http-responses.html
- client side simple parsing of `response multipart/form-data`
  - https://stackoverflow.com/questions/53215506/no-output-after-multipart-newreader
