# End To End Test

To run tests, run the following commands from the project's root directory.

```
make coredns
cd test
./test.sh
```

The `test.sh` script brings up CoreDNS and makes requests with `dig`.
