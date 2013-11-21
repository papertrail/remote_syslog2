loggo - hierarchical loggers for Go
===================================

This package provides an alternative to the standard library log package.

The actual logging functions never return errors.  If you are logging
something, you really don't want to be worried about the logging
having trouble.

Modules have names that are defined by dotted strings.
```
"first.second.third"
```

There is a root module that has the name `""`.  Each module
(except the root module) has a parent, identified by the part of
the name without the last dotted value.
* the parent of `"first.second.third"` is `"first.second"`
* the parent of `"first.second"` is `"first"`
* the parent of `"first"` is `""` (the root module)

Each module can specify its own severity level.  Logging calls that are of
a lower severity than the module's effective severity level are not written
out.

Loggers are created using the GetLogger function.
```
logger := loggo.GetLogger("foo.bar")
```

By default there is one writer registered, which will write to Stderr,
and the root module, which will only emit warnings and above.

Use the usual `Sprintf` syntax for:
```
logger.Criticalf(...)
logger.Errorf(...)
logger.Warningf(...)
logger.Infof(...)
logger.Debugf(...)
logger.Tracef(...)
```

If in doubt, look at the tests. They are quite good at observing the expected behaviour.

The loggers themselves are go-routine safe.  The locking is on the output to the writers, and transparent to the caller.
