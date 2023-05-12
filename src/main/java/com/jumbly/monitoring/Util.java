package com.jumbly.monitoring;

import java.net.InetAddress;
import java.net.UnknownHostException;

public class Util {
    
    public static final String getHostname() {
        String hostname;
        try {
            hostname = InetAddress.getLocalHost().getHostName();
        }
        catch(UnknownHostException e) {
            hostname = "localhost";
        }

        return hostname;
    }

    // TODO(john): get this from external config
    public static final String getDstatHostname() {
        return "localhost";
    }

    // TODO(john): get this from external config
    public static final int getDstatPort() {
        return 8125;
    }
}
