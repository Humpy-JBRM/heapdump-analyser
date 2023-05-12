package com.jumbly.monitoring;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertTrue;

import com.jumbly.monitoring.GCMetricsExporter.ThreadTiming;

import org.junit.Test;

public class GCMetricsExporterTest {
    
    @Test
    public void parseProcLineHappyPathNoSpacesInName() throws Exception {
        String procLine = "31018 (Thread_No_Space) S 13243 30995 13243 34819 30995 1077936192 74300 1235 0 0 1732 17 1 0 20 0 41 0 6444728 22073769984 1995315 18446744073709551615 94328844623872 94328844629157 140726001193872 0 0 0 4 0 16800975 1 0 0 -1 15 0 0 0 0 0 94328844639552 94328844640280 94328863748096 140726001201201 140726001201615 140726001201615 140726001205226 0";
        ThreadTiming timing = ThreadTiming.parseProcLine(procLine);
        assertNotNull("'timing' should not be null", timing);
        assertFalse("timing should not be a GC thread", timing.gcThread);
        assertEquals(String.format("expected ticks = 1749, got %d", timing.ticks), 1749, timing.ticks);
        assertEquals(String.format("expected name = '(Thread_No_Space)', got %s", timing.name), "(Thread_No_Space)", timing.name);
    }

    @Test
    public void parseProcLineHappyPathSpacesInName() throws Exception {
        String procLine = "31018 (Thread With Space) S 13243 30995 13243 34819 30995 1077936192 74300 1235 0 0 1732 17 1 0 20 0 41 0 6444728 22073769984 1995315 18446744073709551615 94328844623872 94328844629157 140726001193872 0 0 0 4 0 16800975 1 0 0 -1 15 0 0 0 0 0 94328844639552 94328844640280 94328863748096 140726001201201 140726001201615 140726001201615 140726001205226 0";
        ThreadTiming timing = ThreadTiming.parseProcLine(procLine);
        assertNotNull("'timing' should not be null", timing);
        assertFalse("timing should not be a GC thread", timing.gcThread);
        assertEquals(String.format("expected ticks = 1749, got %d", timing.ticks), 1749, timing.ticks);
        assertEquals(String.format("expected name = '(Thread With Space)', got %s", timing.name), "(Thread With Space)", timing.name);
    }

    @Test
    public void parseProcLineGCThread() throws Exception {
        String procLine = "31018 (GC Thread#3) S 13243 30995 13243 34819 30995 1077936192 74300 1235 0 0 1732 17 1 0 20 0 41 0 6444728 22073769984 1995315 18446744073709551615 94328844623872 94328844629157 140726001193872 0 0 0 4 0 16800975 1 0 0 -1 15 0 0 0 0 0 94328844639552 94328844640280 94328863748096 140726001201201 140726001201615 140726001201615 140726001205226 0";
        ThreadTiming timing = ThreadTiming.parseProcLine(procLine);
        assertNotNull("'timing' should not be null", timing);
        assertTrue("timing should not be a GC thread", timing.gcThread);
        assertEquals(String.format("expected ticks = 1749, got %d", timing.ticks), 1749, timing.ticks);
        assertEquals(String.format("expected name = '(GC Thread#3)', got %s", timing.name), "(GC Thread#3)", timing.name);
    }

    public void parseProcLineG1Thread() throws Exception {
        String procLine = "31018 (G1 Refine #3) S 13243 30995 13243 34819 30995 1077936192 74300 1235 0 0 1732 17 1 0 20 0 41 0 6444728 22073769984 1995315 18446744073709551615 94328844623872 94328844629157 140726001193872 0 0 0 4 0 16800975 1 0 0 -1 15 0 0 0 0 0 94328844639552 94328844640280 94328863748096 140726001201201 140726001201615 140726001201615 140726001205226 0";
        ThreadTiming timing = ThreadTiming.parseProcLine(procLine);
        assertNotNull("'timing' should not be null", timing);
        assertFalse("timing should not be a GC thread", timing.gcThread);
        assertEquals(String.format("expected ticks = 1749, got %d", timing.ticks), 1749, timing.ticks);
        assertEquals(String.format("expected name = '(G1 Refine #3)', got %s", timing.name), "(G1 Refine #3)", timing.name);
    }

    @Test
    public void parseFoo() throws Exception {
        ThreadTiming[] tt = new GCMetricsExporter().getThreadTimings();
        if (tt.length == 0) {
            
        }
    }
}
