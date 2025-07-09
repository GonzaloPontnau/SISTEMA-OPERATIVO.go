package main

import (
	"sort"
	"time"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

const (
	maxIntentosMemoria     = 5
	tiempoEsperaReintentos = 2 * time.Second
)

// PlanificarLargoPlazo optimizado
func PlanificarLargoPlazo() {
	// Recuperación de pánico para diagnóstico
	defer func() {
		if r := recover(); r != nil {
			utils.ErrorLog.Error("PÁNICO EN PLANIFICADOR DE LARGO PLAZO", "error", r)
			// Re-panic para terminar el programa y mostrar el stack trace
			panic(r)
		}
	}()

	utils.InfoLog.Info("LTS: Iniciando Planificador de Largo Plazo...")

	for {
		var pcb *PCB

		// Prioridad: SUSP. READY > NEW
		suspReadyMutex.Lock()
		if len(colaSuspReady) > 0 {
			pcb = colaSuspReady[0]
			colaSuspReady = colaSuspReady[1:]
			suspReadyMutex.Unlock()

			// Para procesos en SUSP.READY, necesitamos espacio en memoria
			semaforoMultiprogram.Wait()

			if notificarDesswapAMemoria(pcb.PID) {
				pcb.CambiarEstado(EstadoReady)
				readyMutex.Lock()
				colaReady = append(colaReady, pcb)
				readyMutex.Unlock()
				condReady.Signal()
				utils.InfoLog.Info("LTS: Señal enviada al STS para proceso", "pid", pcb.PID)
				continue
			} else {
				// Si falla el desswap, devolver el proceso a SUSP.READY y liberar semáforo
				suspReadyMutex.Lock()
				colaSuspReady = append(colaSuspReady, pcb)
				suspReadyMutex.Unlock()
				semaforoMultiprogram.Signal()
				continue
			}
		}
		suspReadyMutex.Unlock()

		// Procesar cola NEW
		newMutex.Lock()
		for len(colaNew) == 0 {
			// Si no hay procesos en NEW, esperar hasta que llegue uno
			condNew.Wait()
		}

		pcb = seleccionarProcesoLTS()
		if pcb == nil {
			newMutex.Unlock()
			continue
		}
		newMutex.Unlock()

		// --- CASO ESPECIAL PARA EL PROCESO INICIAL (PID 0) ---
		if pcb.PID == 0 {
			utils.InfoLog.Info("LTS: Admitiendo proceso inicial (PID 0) sin consultar memoria.")
			utils.InfoLog.Debug("LTS: Removiendo proceso inicial de cola NEW...")
			// El newMutex ya está bloqueado en este contexto, usar directamente removerDeCola
			removerDeCola(&colaNew, pcb)
			utils.InfoLog.Debug("LTS: Proceso inicial removido de NEW exitosamente")
			pcb.CambiarEstado(EstadoReady)
			readyMutex.Lock()
			colaReady = append(colaReady, pcb)
			readyMutex.Unlock()
			condReady.Signal()
			utils.InfoLog.Info("LTS: Señal enviada al STS para proceso inicial", "pid", pcb.PID)
			continue // Volver al inicio del bucle para procesar el siguiente
		}
		// --- FIN CASO ESPECIAL ---

		// AHORA esperamos el semáforo antes de intentar inicializar en memoria
		semaforoMultiprogram.Wait()

		// Intentos de inicialización en memoria
		if inicializarEnMemoriaConReintentos(pcb) {
			removerDeCola(&colaNew, pcb)
			pcb.CambiarEstado(EstadoReady)

			readyMutex.Lock()
			colaReady = append(colaReady, pcb)
			readyMutex.Unlock()
			condReady.Signal()
		} else {
			// Si falla la inicialización, liberar el semáforo y finalizar proceso
			removerDeCola(&colaNew, pcb)
			FinalizarProceso(pcb, "ERROR_INICIALIZACION_MEMORIA")
			semaforoMultiprogram.Signal()
		}
	}
}

// inicializarEnMemoriaConReintentos maneja reintentos automáticamente
func inicializarEnMemoriaConReintentos(pcb *PCB) bool {
	for intento := 1; intento <= maxIntentosMemoria; intento++ {
		if inicializarProcesoEnMemoria(pcb.PID, pcb.Tamanio, pcb.NombreArchivo) {
			return true
		}

		if intento < maxIntentosMemoria {
			time.Sleep(tiempoEsperaReintentos)
		}
	}
	return false
}

// seleccionarProcesoLTS selecciona el próximo proceso de cola NEW según algoritmo configurado
func seleccionarProcesoLTS() *PCB {
	if len(colaNew) == 0 {
		return nil
	}

	algoritmo := kernelConfig.ReadyIngressAlgorithm
	utils.InfoLog.Debug("LTS seleccionando proceso", "algoritmo", algoritmo, "procesos_disponibles", len(colaNew))

	switch algoritmo {
	case "FIFO":
		return seleccionarFIFOLTS()
	case "PMCP":
		return seleccionarPMCP()
	default:
		utils.InfoLog.Warn("Algoritmo LTS no reconocido, usando FIFO", "algoritmo", algoritmo)
		return seleccionarFIFOLTS()
	}
}

// seleccionarFIFOLTS implementa selección FIFO para LTS
func seleccionarFIFOLTS() *PCB {
	return colaNew[0]
}

// seleccionarPMCP implementa Programación Multiprogramada Controlada por Prioridad
// Prioriza procesos más pequeños para maximizar la multiprogramación
func seleccionarPMCP() *PCB {
	if len(colaNew) == 0 {
		return nil
	}

	// Crear copia para ordenar sin modificar la cola original
	candidatos := make([]*PCB, len(colaNew))
	copy(candidatos, colaNew)

	// Ordenar por tamaño (menor tamaño = mayor prioridad para maximizar multiprogramación)
	sort.Slice(candidatos, func(i, j int) bool {
		// Si tienen el mismo tamaño, usar FIFO (orden de llegada)
		if candidatos[i].Tamanio == candidatos[j].Tamanio {
			return candidatos[i].HoraCreacion.Before(candidatos[j].HoraCreacion)
		}
		return candidatos[i].Tamanio < candidatos[j].Tamanio
	})

	seleccionado := candidatos[0]
	utils.InfoLog.Debug("PMCP seleccionó proceso",
		"pid", seleccionado.PID,
		"tamaño", seleccionado.Tamanio,
		"hora_creacion", seleccionado.HoraCreacion.Format("15:04:05.000"))

	return seleccionado
}

// inicializarProcesoEnMemoria simplificado
func inicializarProcesoEnMemoria(pid int, tamanio int, nombreArchivo string) bool {
	datos := map[string]interface{}{
		"pid":       pid,
		"tamanio":   tamanio,
		"archivo":   nombreArchivo,
		"operacion": "INICIALIZAR_PROCESO",
	}

	respuesta, err := memoriaClient.EnviarHTTPOperacion("INICIALIZAR_PROCESO", datos)
	if err != nil {
		utils.ErrorLog.Error("Error al inicializar proceso en Memoria", "pid", pid, "error", err)
		return false
	}

	if respuestaMap, ok := respuesta.(map[string]interface{}); ok {
		status, _ := respuestaMap["status"].(string)
		return status == "OK"
	}

	return false
}

// notificarDesswapAMemoria notifica a memoria que debe cargar un proceso desde swap
func notificarDesswapAMemoria(pid int) bool {
	datos := map[string]interface{}{
		"pid":       pid,
		"operacion": "DESUSPENDER_PROCESO",
	}

	respuesta, err := memoriaClient.EnviarHTTPOperacion("DESUSPENDER_PROCESO", datos)
	if err != nil {
		return false
	}

	if respuestaMap, ok := respuesta.(map[string]interface{}); ok {
		status, _ := respuestaMap["status"].(string)
		return status == "OK"
	}

	return false
}
