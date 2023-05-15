
# JVM Glass Box
Making JVMs more transparent:

  - automatically take periodic heapdumps when memory use exceeds a given threshold
  
  - export thread states (on a per-thread and aggregated count basis) to Datadog (see *Thread Metrics* below)

  - export detailed GC information to Datadog (see *GC Metrics* below)

## QuickStart - Hoover Up OOM Heapdumps
To do this, all you need to do is:

  - add a `emptyDir{}` volume to your pod `yaml`

  - add the `volumeMount` to your existing docker container to your pod `yaml`
  
  - add the sidecar to your pod `yaml` (See [heapy-demo.yml](heapy-demo.yml))
  
  - Make sure that your VM is running with `-XX:+HeapDumpOnOutOfMemoryError` and `-XX:HeapDumpPath=/heapdumps` as VM parameters
  
  - Make sure that your VM is running with the VM parameters:
  
      - `-DenableAutoHeapdump=true`
      - `-DheapdumpOnMemoryPercent=NNN`
      - `-DheapdumpDirectory=/heapdumps`
      - `-DheapdumpSleepTime=MMM`
  
Where:

  - `NNN` is the % threshold of memory use at which you want to start taking heapdumps
  
  - `MMM` is the number of milliseconds to sleep between checking if a new heapdump is required (and creating it if (ram_used > NNN%)

## How It Works

All of the moving parts are contained in a single java class in `src/HeapdumpDemo.java`.  This is trivially usable from Scala or java.

This class runs in a separate thread and periodically interrogates the JVM for ram usage values.  If the amount of ram used exceeds a given percentage, a heap dump is automatically written to the directory specified by `-DheapdumpDirectory=/path/to/heapdumps`

The reason why this is useful is because, in the case of a OOM in a Kubernetes environment, you may not be able to get the heapdump produced by `-XX:HeapDumpOnOutOfMemoryError` before the container is recycled by kubernetes.  Yet this heapdump contains all of the information required to diagnose the memory use pathology.

By automatically grabbing heapdumps when ram usage exceeds a configurable percentage, you can buy ourselves enough time to:

- produce the heapdump

- copy the resulting `.hprof` file to an external location (e.g. an S3 bucket)

and do so whether the JVM ultimately recovers or ultimately crashes with a OOM.

We use the *sidecar* technique, where two containers run inside a single pod and share a `volumeMount`:

    spec:
      volumes:
        - name: heapdumps
          emptyDir: {}
      containers:
        - name: heapy
          image: .../heapy:latest
          volumeMounts:
            - name: heapdumps
              mountPath: /heapdumps
        - name: heapdump-hoover
          image: .../monitor:latest
          volumeMounts:
            - name: heapdumps
              mountPath: /heapdumps

The `heapdump-hoover` container is the one that does the heavy lifting of detecting new heapdumps and then shipping them off to S3.  It uses the `inotifywait` utility and `aws-cli` to do this, because getting `s3fs` or `rclone` to work inside a docker container (which both need fuse) needs containers to run with elevated permissions (and it's really difficult to build the containers).

## Exported Metrics

### Thread Metrics

  - `jvm_thread.gauge`
    Per-thread information, so that you can identify threads which are in a particular state.

    Attributes: `name` and `state`

  - `jvm_thread_state.gauge`
    The number of threads in a particular state.  This will be useful for diagnosing threadpool starvation.

    Attributes: `state`

  - `jvm_thread_cpu.counter`
    The amount of cpu (clock ticks, `_SC_CLK_TICK`) consumed by this thread.  This will be useful for identifying which threads are dominating the CPU (e.g. it makes it easier to spot runaway GC)

    Attributes: `name`, `state` and `gc`.  The `gc` attribute is `true` if the thread is a GC thread, otherwise `false`.  This allows, for instance, useful graphs showing the ratio of CPU being used by `gc:non_gc`.

### GC Metrics

### AWS Authentication
Authentication is done in one of two ways:

  - making a `.json` available and pointing the `AWS_WEB_IDENTITY_TOKEN_FILE` environment variable at it
  
  - using the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables

Of these, `AWS_WEB_IDENTITY_TOKEN_FILE` is given priority, but one or other must be used.

## Building The Demo

The `src/Main.java` file in the demo contains a really simple bit of Java which kicks off monitor thread and then goes into a tight loop of adding strings to a list (which will eventually result in a OOM).

You will see from the output that, even though it ultimately blows up with a OOM, you collected a whole bundle of heapdumps before this happened - allowing diagnosis of the pattern of memory use leading up to the OOM.

To build the executable jar, simply:

- `make java`

To build the monitor, simply:

  - `make go`

## Pushing to Docker

To build and push the docker containers, simply:  

- `make docker push`

## Adding to an Existing Docker Container

Simply add the sidecar container and volume mounts to the pod yaml:

    containers:
      - name: heapy
        ...
        volumeMounts:
          - name: heapdumps
            mountPath: /heapdumps
      - name: heapdump-hoover
        ...
        volumeMounts:
          - name: heapdumps
          mountPath: /heapdumps

and make sure that you instantiate the `HeapdumpDaemon` class (the thread will auto start).

  

## Configuration

There are four JVM properties that affect the behaviour of the heap dump daemon, of which three are mandatory:

- `enableAutoHeapdump` (optional)

If this is set to `true`, then the heap dump daemon is enabled.  Any other value, or no value at all, will disable the heap dump daemon

- `heapdumpOnMemoryPercent` (mandatory when `enableAutoHeapdump=true`)

If JVM memory use exceeds this percentage of total JVM memory, then you want to take a heap dump.  This is evaluated every time through the loop (see `heapdumpSleepTime` below)

- `heapdumpDirectory` (mandatory when `enableAutoHeapdump=true`)

Heap dumps will be written to this directory, which must already exist.  The files will be named `<unix_time>.hprof`

- `heapdumpSleepTime` (mandatory when `enableAutoHeapdump=true`)

The frequency that the daemon will check the memory use, in milliseconds.  Setting this too low will cause a lot of heap dumps to be created.  Setting this too high may cause the daemon to miss a memory spike which happens in a very small timeframe.
  
## Automated Heapdump Analysis
Analysing heapdumps with [MAT](https://www.eclipse.org/mat/) can be tedious (but also fun - I hjghly recommend it).

If there are a lot of heapdumps to analyse, it's time-consuming - you might want the various reports to be auto-generated so that you can speed up getting to the root of the problem.

To do this, just add the `-analyse` flag to the command in the `Dockerfile`:

  - `CMD ["/main", "-analyse", "/heapdumps"]`

To disable automated heapdump analysis, set the flag to false (or omit it):

  - `CMD ["/main", "-analyse=false", "/heapdumps"]`

Every heapdump produced will then have three zip files:

  - `heapdump_Leak_Suspects.zip` (the Leak Suspects report)
  - `heapdump_Top_Components.zip ` (the Top Components report)
  - `heapdump_System_Overview.zip` (the VM summary report)

unzipping these and opening the `index.html` in a browser will show exactly the same reports you get from MAT:

  - [Sample](sample-reports/summary/index.html)
  - [Leak Suspects](sample-reports/leak-suspects/index.html)

Detailed information on what these reports contain and how to decipher them is available at [MAT Documentation](https://help.eclipse.org/latest/index.jsp?topic=/org.eclipse.mat.ui.help/welcome.html).

You can access then in the S3 bucket with `aws s3`:

```
$ aws s3 ls jumbly-s3spike --recursive
2022-05-18 10:59:07    1595928 heapy-demo/hprof/heapdump.a2s.index
2022-05-18 10:59:07    3171275 heapy-demo/hprof/heapdump.domIn.index
2022-05-18 10:59:07    9795777 heapy-demo/hprof/heapdump.domOut.index
2022-05-18 10:59:07   58578221 heapy-demo/hprof/heapdump.hprof
2022-05-18 10:59:07      19980 heapy-demo/hprof/heapdump.i2sv2.index
2022-05-18 10:59:07    9574835 heapy-demo/hprof/heapdump.idx.index
2022-05-18 10:59:12   11215064 heapy-demo/hprof/heapdump.inbound.index
2022-05-18 10:59:12     679343 heapy-demo/hprof/heapdump.index
2022-05-18 10:59:13    2825384 heapy-demo/hprof/heapdump.o2c.index
2022-05-18 10:59:14    9041444 heapy-demo/hprof/heapdump.o2hprof.index
2022-05-18 10:59:14    6838241 heapy-demo/hprof/heapdump.o2ret.index
2022-05-18 10:59:14   11090064 heapy-demo/hprof/heapdump.outbound.index
2022-05-18 10:59:17      10197 heapy-demo/hprof/heapdump.threads
2022-05-18 10:59:16      77986 heapy-demo/hprof/heapdump_Leak_Suspects.zip
2022-05-18 10:59:16      55784 heapy-demo/hprof/heapdump_System_Overview.zip
2022-05-18 10:59:17     122197 heapy-demo/hprof/heapdump_Top_Components.zip
```

(TODO(john): expand the zip files in s3: and serve them from there - plus the .hprof too)

## Housekeeping

The monitor container will remove heap dumps from the filesystem when they have been copied to the S3 bucket, to make sure that you don't completely fill up the container filesystem.

It's probably a good idea to size the container filesystems to at least `(JVM_ram_size * 2) + whatever_else_you_need`  to allow enough disk space for at least two heap dumps.

The naming of the `.hprof` files - with a unix timestamp - allows for easy housekeeping of the resulting heap dumps in the S3 bucket because it's easy to identify old ones simply from their filename alone.

The `s3:` bucket will need housekeeping - probably best done simply by setting the retention time to a short value (e.g. a few days).
