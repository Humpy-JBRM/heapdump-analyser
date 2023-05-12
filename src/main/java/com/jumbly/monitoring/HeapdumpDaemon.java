package com.jumbly.monitoring;

import java.io.IOException;
import com.sun.management.HotSpotDiagnosticMXBean;
import javax.management.MBeanServer;
import java.lang.management.ManagementFactory;

/**
 * HeapDumpDaemon is a worker thread which will create a heapdump
 * when the memory usage of the VM exceeds a certain threshold.
 * <p/>
 * This is useful for automatically generating heapdumps from VMs
 * which are displaying pathological ram usage intermittently so
 * that the dump can be analysed to identify the root cause of the
 * problem.
 * <p/>
 * To use (Java):
 * <p/>
 * <pre>
 *    try {
 *        HeapdumpDaemon hdd = new HeapdumpDaemon();
 *        synchronized(hdd) {
 *            hdd.wait();
 *        }
 *    }
 *    catch(InterruptedException e) {
 *    }
 * </pre>
 * To use (Scala):
 * <p/>
 * <pre>
 *     val hdd = new HeapdumpDaemon();
 * </pre>
 * 
 * The daemon thread will auto-start.
 */
public final class HeapdumpDaemon extends Thread {
    static final String PROP_ENABLE_HEAP_DUMP_DAEMON = "enableAutoHeapdump";
    static final String PROP_HEAP_DUMP_MEMORY_PERCENT = "heapdumpOnMemoryPercent";
    static final String PROP_HEAP_DUMP_DIRECTORY = "heapdumpDirectory";
    static final String PROP_HEAP_DUMP_SLEEP_TIME = "heapdumpSleepTime";

    long totalMemory;
    long memoryPercent;
    long sleepTimeMillis;
    long heapDumpIfWeUseTheseManyBytes;
    String heapdumpDir;

    public HeapdumpDaemon() {
        super();

        // Validate the parameters
        if (System.getProperty(PROP_ENABLE_HEAP_DUMP_DAEMON) != null && System.getProperty(PROP_ENABLE_HEAP_DUMP_DAEMON).equalsIgnoreCase("true")) {
            if (System.getProperty(PROP_HEAP_DUMP_MEMORY_PERCENT) == null || System.getProperty(PROP_HEAP_DUMP_MEMORY_PERCENT).trim() == "") {
                throw new IllegalArgumentException(String.format("Property '%s' must be set when %s=true", PROP_HEAP_DUMP_MEMORY_PERCENT, PROP_ENABLE_HEAP_DUMP_DAEMON));
            }
            if (System.getProperty(PROP_HEAP_DUMP_DIRECTORY) == null || System.getProperty(PROP_HEAP_DUMP_DIRECTORY).trim() == "") {
                throw new IllegalArgumentException(String.format("Property '%s' must be set when %s=true", PROP_HEAP_DUMP_DIRECTORY, PROP_ENABLE_HEAP_DUMP_DAEMON));
            }
            if (System.getProperty(PROP_HEAP_DUMP_SLEEP_TIME) == null || System.getProperty(PROP_HEAP_DUMP_SLEEP_TIME).trim() == "") {
                throw new IllegalArgumentException(String.format("Property '%s' must be set when %s=true", PROP_HEAP_DUMP_SLEEP_TIME, PROP_ENABLE_HEAP_DUMP_DAEMON));
            }
            
            totalMemory = Runtime.getRuntime().totalMemory();
            System.out.printf("totalMemory = %d\n", totalMemory);
            try {
                memoryPercent = Long.parseLong(System.getProperty(PROP_HEAP_DUMP_MEMORY_PERCENT).trim());
            }
            catch(Exception e) {
                throw new IllegalArgumentException(String.format("Property '%s' value must be a number, not '%s'", PROP_HEAP_DUMP_MEMORY_PERCENT, System.getProperty(PROP_HEAP_DUMP_MEMORY_PERCENT).trim()));
            }

            heapDumpIfWeUseTheseManyBytes = (memoryPercent * totalMemory) / 100;
            System.out.printf("heapDumpIfWeUseTheseManyBytes = %d\n", heapDumpIfWeUseTheseManyBytes);
            try {
                sleepTimeMillis = Long.parseLong(System.getProperty(PROP_HEAP_DUMP_SLEEP_TIME).trim());
            }
            catch(Exception e) {
                throw new IllegalArgumentException(String.format("Property '%s' value must be a number, not '%s'", PROP_HEAP_DUMP_SLEEP_TIME, System.getProperty(PROP_HEAP_DUMP_SLEEP_TIME).trim()));
            }

            heapdumpDir = System.getProperty(PROP_HEAP_DUMP_DIRECTORY).trim();
        }

        this.setDaemon(true);
        this.setPriority(NORM_PRIORITY);
        this.start();
    }

    public static void dumpHeap(String filePath, boolean live) {
        try {
            System.out.printf("Dumping to %s\n", filePath);
            MBeanServer server = ManagementFactory.getPlatformMBeanServer();
            HotSpotDiagnosticMXBean mxBean = ManagementFactory.newPlatformMXBeanProxy(
            server, "com.sun.management:type=HotSpotDiagnostic", HotSpotDiagnosticMXBean.class);
            mxBean.dumpHeap(filePath, live);

            // TODO(john): Thread dump
            // ThreadInfo[] allThreads = ManagementFactory.getThreadMXBean().dumpAllThreads(true, true);

            // TODO(john): GC info
            // ManagementFactory.getGarbageCollectorMXBeans()...
        } catch (IOException e) {
            System.err.println("HeapdumpDaemon.dumpHeap(): Could not create heap dump");
            e.printStackTrace();
        }
    }

    String humanise(long valueInBytes) {
        if (valueInBytes >= 1073741824) {
            return String.format("%.2fG", (double) valueInBytes / 1073741824D);
        }
        if (valueInBytes >= 1048576) {
            return String.format("%.2fM", (double) valueInBytes / 1048576D);
        }
        if (valueInBytes >= 1024) {
            return String.format("%.2fM", (double) valueInBytes / 1024D);
        }
        return String.format("%d", valueInBytes);
    }

    @Override
    public void run() {
        if (System.getProperty(PROP_ENABLE_HEAP_DUMP_DAEMON) == null || !System.getProperty(PROP_ENABLE_HEAP_DUMP_DAEMON).equalsIgnoreCase("true")) {
            System.out.printf("Not starting HeapdumpDaemon because property '%s' is not set to 'true'\n", PROP_ENABLE_HEAP_DUMP_DAEMON);
            return;
        }
        System.out.printf("Starting HeapdumpDaemon: heap dump to %s when memory is %d%% (%s / %s)\n", this.heapdumpDir, this.memoryPercent, humanise(this.heapDumpIfWeUseTheseManyBytes), humanise(this.totalMemory));
        while (true) {
            long usedMemory = Runtime.getRuntime().totalMemory() - Runtime.getRuntime().freeMemory();
            if (usedMemory > heapDumpIfWeUseTheseManyBytes) {
                System.out.printf("Dumping heap because used memory (%d) > %d\n", usedMemory, heapDumpIfWeUseTheseManyBytes);
                dumpHeap(String.format("%s/%d.hprof", heapdumpDir, System.currentTimeMillis() / 1000L), true);
            }
            try {
                Thread.sleep(sleepTimeMillis);
            } catch (Exception e) {
            }
        }
    }
}
