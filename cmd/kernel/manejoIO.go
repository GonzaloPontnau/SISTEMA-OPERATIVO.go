package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

var (
	dispositivosIO      map[string]*utils.HTTPClient = make(map[string]*utils.HTTPClient)
	dispositivosIOMutex sync.RWMutex
	contadorBalanceador int // Contador para round-robin
	balanceadorMutex    sync.Mutex
)

// RegistrarDispositivoIO optimizado
func RegistrarDispositivoIO(nombre string, ip string, puerto int) {
	dispositivosIOMutex.Lock()
	defer dispositivosIOMutex.Unlock()

	if _, existe := dispositivosIO[nombre]; !existe {
		dispositivosIO[nombre] = utils.NewHTTPClient(ip, puerto, "Kernel->"+nombre)
		utils.InfoLog.Info("Dispositivo IO registrado", "nombre", nombre, "ip", ip, "puerto", puerto)
	}
}

func ObtenerClienteIO(nombre string) (*utils.HTTPClient, bool) {
	dispositivosIOMutex.RLock()
	defer dispositivosIOMutex.RUnlock()
	cliente, existe := dispositivosIO[nombre]
	return cliente, existe
}

// EnviarSolicitudIO simplificado
func EnviarSolicitudIO(pcb *PCB, dispositivo string, tiempo int) {
	cliente, existe := ObtenerClienteIO(dispositivo)
	if !existe {
		utils.ErrorLog.Error("Dispositivo IO no registrado", "dispositivo", dispositivo, "pid", pcb.PID)
		utils.InfoLog.Info(fmt.Sprintf("## (%d) - IO no disponible, continuando ejecución", pcb.PID))
		// En lugar de finalizar, incrementar PC y volver a READY para continuar el bucle infinito
		pcb.PC++
		MoverProcesoAReady(pcb)
		return
	}

	utils.InfoLog.Info(fmt.Sprintf("## (%d) - Enviando petición a IO: %s", pcb.PID, dispositivo))

	datos := map[string]interface{}{
		"pid":       pcb.PID,
		"tiempo":    tiempo,
		"operacion": "IO_REQUEST",
	}

	_, err := cliente.EnviarHTTPOperacion("IO_REQUEST", datos)
	if err != nil {
		utils.ErrorLog.Error("Error comunicación IO", "dispositivo", dispositivo, "pid", pcb.PID, "error", err.Error())
		utils.InfoLog.Info(fmt.Sprintf("## (%d) - Error de comunicación IO, continuando ejecución", pcb.PID))
		// En lugar de finalizar, incrementar PC y volver a READY para continuar el bucle infinito
		pcb.PC++
		MoverProcesoAReady(pcb)
	}
}

// ManejadorRegistroIO simplificado
func ManejadorRegistroIO(origen string, datos map[string]interface{}) (interface{}, bool) {
	tipoModulo, ok := datos["tipo"].(string)
	if !ok || !strings.HasPrefix(tipoModulo, "IO") {
		return nil, false
	}

	ip, okIP := datos["ip"].(string)
	puertoFloat, okPuerto := datos["puerto"].(float64)

	if !okIP || !okPuerto {
		return map[string]interface{}{
			"status":  "ERROR",
			"message": "Handshake incompleto",
		}, true
	}

	// Registrar dispositivo con nombre completo (IODISCO, IODISCO2)
	RegistrarDispositivoIO(tipoModulo, ip, int(puertoFloat))

	// También registrar con nombre simplificado (DISCO, DISCO2) para compatibilidad
	nombreSimplificado := strings.TrimPrefix(tipoModulo, "IO")
	if nombreSimplificado != tipoModulo {
		RegistrarDispositivoIO(nombreSimplificado, ip, int(puertoFloat))
		utils.InfoLog.Info(fmt.Sprintf("## Dispositivo IO registrado: '%s' y '%s'", tipoModulo, nombreSimplificado))
	} else {
		utils.InfoLog.Info(fmt.Sprintf("## Módulo IO '%s' registrado", tipoModulo))
	}

	return map[string]interface{}{
		"status":  "OK",
		"message": fmt.Sprintf("IO '%s' registrado", tipoModulo),
	}, true
}

// ProcesarSolicitudIO optimizado
func ProcesarSolicitudIO(datos map[string]interface{}) (interface{}, bool) {
	evento, _ := datos["evento"].(string)
	motivo, _ := datos["motivo_retorno"].(string)

	if evento != "SOLICITUD_IO" && motivo != "IO_REQUEST" {
		return nil, false
	}

	pidFloat, pidOk := datos["pid"].(float64)
	if !pidOk {
		return map[string]interface{}{"status": "ERROR", "mensaje": "PID inválido"}, true
	}

	dispositivo, _ := datos["dispositivo"].(string)
	if dispositivo == "" {
		dispositivo, _ = datos["nombre_dispositivo"].(string)
	}

	tiempoFloat, _ := datos["tiempo"].(float64)
	if tiempoFloat == 0 {
		tiempoFloat, _ = datos["tiempo_bloqueo"].(float64)
	}

	pcb := BuscarPCBPorPID(int(pidFloat))
	if pcb == nil {
		return map[string]interface{}{"status": "ERROR", "mensaje": "Proceso no encontrado"}, true
	}

	dispositivoSeleccionado := SeleccionarDispositivoIO(dispositivo, pcb.PID)
	MoverProcesoABlocked(pcb, fmt.Sprintf("IO_%s", dispositivoSeleccionado))
	go EnviarSolicitudIO(pcb, dispositivoSeleccionado, int(tiempoFloat))
	go despacharProcesoSiCorresponde()

	return map[string]interface{}{"status": "OK", "mensaje": "IO procesando"}, true
}

// ProcesarIOTerminada optimizado
func ProcesarIOTerminada(datos map[string]interface{}) (interface{}, bool) {
	evento, _ := datos["evento"].(string)
	operacion, _ := datos["operacion"].(string)

	if evento != "IO_TERMINADA" && operacion != "IO_COMPLETADA" {
		return nil, false
	}

	pidFloat, pidOk := datos["pid"].(float64)
	if !pidOk {
		return map[string]interface{}{"status": "ERROR", "mensaje": "PID inválido"}, true
	}

	pcb := BuscarPCBPorPID(int(pidFloat))
	if pcb == nil || pcb.Estado != EstadoBlocked {
		return map[string]interface{}{"status": "WARN", "mensaje": "Proceso no válido"}, true
	}

	utils.InfoLog.Info(fmt.Sprintf("## (%d) finalizó IO y pasa a READY", pcb.PID))
	// Incrementar PC para pasar a la siguiente instrucción después de la I/O
	pcb.PC++
	MoverProcesoAReady(pcb)
	go despacharProcesoSiCorresponde()

	return map[string]interface{}{"status": "OK", "mensaje": "IO completada"}, true
}

// SeleccionarDispositivoIO implementa balanceador de carga genérico para dispositivos IO
func SeleccionarDispositivoIO(dispositivoSolicitado string, pid int) string {
	dispositivosIOMutex.RLock()
	defer dispositivosIOMutex.RUnlock()

	// Si el dispositivo solicitado existe directamente, usarlo
	if _, existe := dispositivosIO[dispositivoSolicitado]; existe {
		utils.InfoLog.Info(fmt.Sprintf("## (%d) - Usando dispositivo directo: %s", pid, dispositivoSolicitado))
		return dispositivoSolicitado
	}

	// Si el dispositivo no existe, buscar dispositivos similares para distribución automática
	dispositivosSimilares := obtenerDispositivosSimilares(dispositivoSolicitado)
	if len(dispositivosSimilares) == 0 {
		utils.ErrorLog.Error("No hay dispositivos disponibles", "dispositivo_solicitado", dispositivoSolicitado, "pid", pid)
		return dispositivoSolicitado // Fallback al original
	}

	// Usar round-robin para distribuir carga entre dispositivos disponibles
	balanceadorMutex.Lock()
	dispositivoSeleccionado := dispositivosSimilares[contadorBalanceador%len(dispositivosSimilares)]
	contadorBalanceador++
	balanceadorMutex.Unlock()

	utils.InfoLog.Info(fmt.Sprintf("## (%d) - Balanceador IO: %s -> %s", pid, dispositivoSolicitado, dispositivoSeleccionado))
	return dispositivoSeleccionado
}

// obtenerDispositivosSimilares devuelve lista de todos los dispositivos IO disponibles
func obtenerDispositivosSimilares(dispositivoBase string) []string {
	var dispositivos []string

	// Obtener todos los dispositivos IO registrados
	for nombre := range dispositivosIO {
		dispositivos = append(dispositivos, nombre)
	}

	// Ordenar para consistencia en el balanceador
	sort.Strings(dispositivos)
	return dispositivos
}
