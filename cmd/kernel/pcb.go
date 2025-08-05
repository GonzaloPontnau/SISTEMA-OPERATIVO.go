package main

import (
	"fmt"
	"time"

	"github.com/GonzaloPontnau/SISTEMA-OPERATIVO.go.git/utils"
)

const (
	EstadoNew         = "NEW"
	EstadoReady       = "READY"
	EstadoExec        = "EXEC"
	EstadoBlocked     = "BLOCKED"
	EstadoSuspReady   = "SUSP. READY"
	EstadoSuspBlocked = "SUSP. BLOCKED"
	EstadoExit        = "EXIT"
)

type PCB struct {
	PID                       int
	Estado                    string
	NombreArchivo             string
	Tamanio                   int
	PC                        int
	EstimacionSiguienteRafaga float64

	// Timestamps
	HoraCreacion     time.Time
	HoraListo        time.Time
	HoraEjecucion    time.Time
	HoraBloqueo      time.Time
	HoraFinalizacion time.Time

	// Tracking de ejecución
	UltimaRafagaReal     float64
	InicioUltimaRafaga   time.Time
	TotalEjecuciones     int
	TotalTiempoEjecucion float64
	MotivoBloqueo        string

	// Tracking de estados para métricas
	TotalReady        int
	TotalTiempoReady  float64
	InicioUltimoReady time.Time

	// Flag para distinguir si el proceso está realmente en SWAP o ya fue cargado por IO
	EnSwap bool
}

// NuevoPCB simplificado
func NuevoPCB(pid int, tamanio int) *PCB {
	horaActual := time.Now()
	finalPID := pid
	if pid < 0 {
		finalPID = GenerarNuevoPID()
	}

	estimacionInicial := float64(kernelConfig.InitialEstimate)
	if estimacionInicial <= 0 {
		estimacionInicial = 5000.0
	}

	pcb := &PCB{
		PID:                       finalPID,
		Estado:                    EstadoNew,
		Tamanio:                   tamanio,
		PC:                        0,
		EstimacionSiguienteRafaga: estimacionInicial,
		HoraCreacion:              horaActual,
		EnSwap:                    false, // Los procesos nuevos no están en SWAP
	}

	mapaMutex.Lock()
	mapaPCBs[pcb.PID] = pcb
	mapaMutex.Unlock()

	utils.InfoLog.Info(fmt.Sprintf("(%d) - Se crea el proceso - Estado: %s", pcb.PID, pcb.Estado))

	return pcb
}

// CambiarEstado optimizado
func (pcb *PCB) CambiarEstado(nuevoEstado string) {
	if pcb.Estado == nuevoEstado {
		return
	}

	estadoAnterior := pcb.Estado
	horaActual := time.Now()

	if estadoAnterior == EstadoReady && !pcb.InicioUltimoReady.IsZero() {
		tiempoEnReady := horaActual.Sub(pcb.InicioUltimoReady).Seconds()
		pcb.TotalTiempoReady += tiempoEnReady
	}

	// Manejar transiciones de ejecución
	switch {
	case estadoAnterior == EstadoReady && nuevoEstado == EstadoExec:
		pcb.InicioUltimaRafaga = horaActual
		pcb.HoraEjecucion = horaActual

	case estadoAnterior == EstadoExec:
		if !pcb.InicioUltimaRafaga.IsZero() {
			pcb.UltimaRafagaReal = horaActual.Sub(pcb.InicioUltimaRafaga).Seconds() * 1000
			pcb.TotalEjecuciones++
			pcb.TotalTiempoEjecucion += pcb.UltimaRafagaReal
			pcb.actualizarEstimacion()
		}
	}

	// Actualizar timestamps y contadores
	switch nuevoEstado {
	case EstadoReady:
		pcb.HoraListo = horaActual
		pcb.InicioUltimoReady = horaActual
		pcb.TotalReady++
	case EstadoBlocked:
		pcb.HoraBloqueo = horaActual
	case EstadoExit:
		pcb.HoraFinalizacion = horaActual
	}

	pcb.Estado = nuevoEstado
	utils.InfoLog.Info(fmt.Sprintf("(%d) - Pasa del estado %s al estado %s", pcb.PID, estadoAnterior, nuevoEstado))
}

// actualizarEstimacion simplificada
func (pcb *PCB) actualizarEstimacion() {
	if pcb.UltimaRafagaReal <= 0 {
		return
	}

	alpha := kernelConfig.Alpha
	if alpha < 0 || alpha > 1 {
		alpha = 0.5
	}

	pcb.EstimacionSiguienteRafaga = alpha*pcb.UltimaRafagaReal + (1-alpha)*pcb.EstimacionSiguienteRafaga
}

func (pcb *PCB) String() string {
	return fmt.Sprintf("PCB{PID: %d, Estado: %s, Tamaño: %d, PC: %d}",
		pcb.PID, pcb.Estado, pcb.Tamanio, pcb.PC)
}

// CalcularMetricas optimizado
func (pcb *PCB) CalcularMetricas() {
	// Si el proceso está saliendo de READY al finalizar, acumular el último tiempo
	if pcb.Estado == EstadoReady && !pcb.InicioUltimoReady.IsZero() {
		tiempoEnReady := time.Now().Sub(pcb.InicioUltimoReady).Seconds()
		pcb.TotalTiempoReady += tiempoEnReady
	}

	tiempoNew := 0.0
	if pcb.HoraListo.After(pcb.HoraCreacion) {
		tiempoNew = pcb.HoraListo.Sub(pcb.HoraCreacion).Seconds()
	}

	tiempoReady := pcb.TotalTiempoReady

	tiempoExec := pcb.TotalTiempoEjecucion / 1000.0

	tiempoBlocked := 0.0
	if !pcb.HoraBloqueo.IsZero() && !pcb.HoraFinalizacion.IsZero() {
		tiempoBlocked = pcb.HoraFinalizacion.Sub(pcb.HoraBloqueo).Seconds()
	}

	utils.InfoLog.Info(fmt.Sprintf("(%d) - Métricas de estado: NEW (%d)(%.2f), READY (%d)(%.2f), EXEC (%d)(%.2f), BLOCKED (%d)(%.2f)",
		pcb.PID, 1, tiempoNew, pcb.TotalReady, tiempoReady, pcb.TotalEjecuciones, tiempoExec, 1, tiempoBlocked))
}
