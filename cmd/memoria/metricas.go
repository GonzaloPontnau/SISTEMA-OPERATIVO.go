package main

// Funciones para actualizar métricas

// Actualizar métricas de acceso a tablas de páginas
func actualizarMetricasAccesoTabla(pid int) {
	if _, existe := metricasPorProceso[pid]; !existe {
		metricasPorProceso[pid] = &MetricasProceso{}
	}
	metricasPorProceso[pid].AccesosTablasPaginas++
}

// Actualizar métricas de instrucciones solicitadas
func actualizarMetricasInstruccion(pid int) {
	if _, existe := metricasPorProceso[pid]; !existe {
		metricasPorProceso[pid] = &MetricasProceso{}
	}
	metricasPorProceso[pid].InstruccionesSolicitadas++
}

// Actualizar métricas de bajadas a SWAP
func actualizarMetricasBajadaSwap(pid int) {
	if _, existe := metricasPorProceso[pid]; !existe {
		metricasPorProceso[pid] = &MetricasProceso{}
	}
	metricasPorProceso[pid].BajadasSwap++
}

// Actualizar métricas de subidas a memoria
func actualizarMetricasSubidaMemoria(pid int) {
	if _, existe := metricasPorProceso[pid]; !existe {
		metricasPorProceso[pid] = &MetricasProceso{}
	}
	metricasPorProceso[pid].SubidasMemoria++
}

// Actualizar métricas de lecturas de memoria
func actualizarMetricasLectura(pid int) {
	if _, existe := metricasPorProceso[pid]; !existe {
		metricasPorProceso[pid] = &MetricasProceso{}
	}
	metricasPorProceso[pid].LecturasMemoria++
}

// Actualizar métricas de escrituras en memoria
func actualizarMetricasEscritura(pid int) {
	if _, existe := metricasPorProceso[pid]; !existe {
		metricasPorProceso[pid] = &MetricasProceso{}
	}
	metricasPorProceso[pid].EscriturasMemoria++
}
