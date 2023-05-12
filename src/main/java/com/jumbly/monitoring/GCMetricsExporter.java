package com.jumbly.monitoring;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import com.timgroup.statsd.NonBlockingStatsDClientBuilder;
import com.timgroup.statsd.StatsDClient;

public class GCMetricsExporter extends Thread {

    static final String PROP_ENABLE_GC_METRICS_EXPORTER = "enableGCMetricsExporter";

    public GCMetricsExporter() {
        super("GCMetricsExporter");
        this.setPriority(MIN_PRIORITY);
        this.setDaemon(true);
        this.start();
    }

    @Override
    public void run() {
        if (System.getProperty(PROP_ENABLE_GC_METRICS_EXPORTER) == null || !System.getProperty(PROP_ENABLE_GC_METRICS_EXPORTER).equalsIgnoreCase("true")) {
            System.out.printf("Not starting GCMetricsExporter because property '%s' is not set to 'true'\n", PROP_ENABLE_GC_METRICS_EXPORTER);
            return;
        }

        StatsDClient statsd = new NonBlockingStatsDClientBuilder()
            .prefix("statsd")
            .hostname(Util.getDstatHostname())
            .port(Util.getDstatPort())
            .build();

        System.out.printf("Starting GCMetricsExporter");
        while (true) {
            for (ThreadTiming tt : this.getThreadTimings()) {
                System.out.printf("jvm_thread_cpu.gauge %d state:%s name:%s gc:%s\n", tt.ticks, tt.state, tt.name, tt.gcThread);
                statsd.recordGaugeValue("jvm_thread_cpu.gauge", tt.ticks, new String[]{"state:" + tt.state, "name:" + tt.name, "gc:" + tt.gcThread});
            }

            try {
                Thread.sleep(1000L);
            } catch (Exception e) {
            }
        }
    }

    // Aggregate for holding thread timing information
    static class ThreadTiming {
        String name;
        boolean gcThread;
        String state;
        long ticks;

        // See proc(5)
        //
        // Field    Description
        //     3    State
        //    14    Clock ticks spent in user mode
        //    15    Clock ticks spent in kernel mode
        static ThreadTiming newThreadTiming(File procDir) {
            ThreadTiming timing = null;
            try {
                File statFile = new File(procDir, "stat");
                String procLine = Files.readString(Path.of(statFile.getAbsolutePath())).trim();
                timing = parseProcLine(procLine);
            }
            catch(IOException ignoreForNow) {
                // TODO(john): Should we log here, or might it be too granular for our level of interest?
            }
            finally {
                return timing;
            }
        }

        static ThreadTiming parseProcLine(String procLine) {
            ThreadTiming timing = null;

            if (!procLine.isEmpty()) {
                // Split this into its individual fields, paying particular attention
                // to thread names containing a space.
                //
                // Without space:
                //
                //    ThreadId ThreadName State ...
                //
                // With space:
                //
                //    ThreadId (Thread Name) State ...
                String[] fields = procLine.split("\\s");
                if (fields.length > 15) {
                    // There looks to be enough fields.
                    int nameField = 1;
                    int stateField = 2;
                    int uTicksField = 13;
                    int kTicksField = 14;
                    String threadName = fields[nameField];
                    if (threadName.startsWith("(") && !threadName.endsWith(")")) {
                        // Thread name contains spaces
                        for (nameField++; nameField < fields.length && !threadName.endsWith(")"); nameField++) {
                            threadName = threadName + " " + fields[nameField];

                            // All of the other fields get shifted right
                            stateField++;
                            uTicksField++;
                            kTicksField++;
                        }
                    }

                    if (kTicksField < fields.length) {
                        long uTicks;
                        long kTicks;

                        try {
                            uTicks = Long.parseLong(fields[uTicksField]);
                        }
                        catch(NumberFormatException e) {
                            // We cannot parse this value.
                            // TODO(john): Should we log here, or might it be too granular for our level of interest?
                            return null;
                        }
                        try {
                            kTicks = Long.parseLong(fields[kTicksField]);
                        }
                        catch(NumberFormatException e) {
                            // We cannot parse this value.
                            // TODO(john): Should we log here, or might it be too granular for our level of interest?
                            return null;
                        }

                        // We now have everything we need to create this timing entry.
                        //
                        // This is the only place in this method where a non-null ThreadTiming
                        // can be created.
                        timing = new ThreadTiming();
                        timing.name = threadName;
                        switch (fields[stateField]) {
                            case "R":
                                timing.state = "Running";
                                break;

                            case "S":
                                timing.state = "Sleeping (interruptable)";
                                break;

                            case "D":
                                timing.state = "Sleeping (non-interruptible)";
                                break;

                            case "Z":
                                timing.state = "Zombie";
                                break;

                            default:
                                timing.state = "Other";
                                break;
                        }
                        timing.gcThread = threadName.indexOf("(GC ") == 0 || threadName.indexOf("(G1 ") == 0;
                        timing.ticks = (uTicks + kTicks);
                    }
                }
            }

            return timing;
        }
    }

    ThreadTiming[] getThreadTimings() {
        List<ThreadTiming> timings = new ArrayList<>();
        String procDir = String.format("/proc/%d/task", ProcessHandle.current().pid());
        File threadsDir = new File(procDir);
        if (threadsDir.isDirectory()) {
            for (String threadId : threadsDir.list()) {
                ThreadTiming threadTiming = ThreadTiming.newThreadTiming(new File(threadsDir, threadId));
                if (threadTiming != null) {
                    timings.add(threadTiming);
                }
            }
        }
        return timings.toArray(new ThreadTiming[0]);
    }
}
