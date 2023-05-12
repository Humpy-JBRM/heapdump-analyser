package com.jumbly.monitoring;

import java.util.Date;
import java.util.LinkedList;
import java.util.List;

public class Main {

	public static void main(String[] args) {
		new HeapdumpDaemon();
		new ThreadStateExporter();
		new GCMetricsExporter();
		
		List<String> fillMeUp = new LinkedList<String>();
		while (true) {
			fillMeUp.add(new Date().toString());
		}
	}
}
