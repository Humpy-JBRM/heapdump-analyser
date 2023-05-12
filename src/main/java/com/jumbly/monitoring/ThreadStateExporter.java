package com.jumbly.monitoring;

import java.util.Map;
import java.net.InetAddress;
import java.net.UnknownHostException;
import java.util.HashMap;
import java.util.concurrent.atomic.AtomicInteger;
import com.timgroup.statsd.NonBlockingStatsDClientBuilder;
import com.timgroup.statsd.StatsDClient;

public class ThreadStateExporter extends Thread {

    static final String PROP_ENABLE_THREAD_STATE_EXPORTER = "enableThreadStateExporter";

    public ThreadStateExporter() {
        super("ThreadStateExporter");
        this.setPriority(MIN_PRIORITY);
        this.setDaemon(true);
        this.start();
    }

    @Override
    public void run() {
        if (System.getProperty(PROP_ENABLE_THREAD_STATE_EXPORTER) == null || !System.getProperty(PROP_ENABLE_THREAD_STATE_EXPORTER).equalsIgnoreCase("true")) {
            System.out.printf("Not starting ThreadStateExporter because property '%s' is not set to 'true'\n", PROP_ENABLE_THREAD_STATE_EXPORTER);
            return;
        }

        // TODO(john): externalised config
        StatsDClient statsd = new NonBlockingStatsDClientBuilder()
            .prefix("statsd")
            .hostname(Util.getDstatHostname())
            .port(Util.getDstatPort())
            .build();

        System.out.printf("Starting ThreadStateExporter");
        while (true) {
            Map<String, AtomicInteger> threadCountByState = new HashMap<String, AtomicInteger>();
            for (Thread t : Thread.getAllStackTraces().keySet()) {
                // No string metrics in datadog ...
                String state = String.format("%s", t.getState());
                System.out.printf("jvm_thread.gauge 1 state:%s name:%s\n", state, t.getName());
                statsd.recordGaugeValue("jvm_thread.gauge", 1D, new String[]{"state:" + state, "name:" + t.getName()});

                AtomicInteger count = threadCountByState.get(state);
                if (count == null) {
                    count = new AtomicInteger();
                    threadCountByState.put(state, count);
                }
                count.incrementAndGet();
            }

            for (String key : threadCountByState.keySet()) {
                // TODO(john): remove this bit of debugging
                //System.out.printf("jvm_thread_state.gauge %8d state:%s\n", threadCountByState.get(key).get(), key);
                statsd.recordGaugeValue("jvm_thread_state.gauge", threadCountByState.get(key).get(), new String[]{"state:" + key});
            }

            try {
                Thread.sleep(10000L);
            } catch (Exception e) {
            }
        }
    }
}
