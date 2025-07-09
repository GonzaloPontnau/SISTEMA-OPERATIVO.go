// main.go (Memoria)
package main

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

var (
	modulo     *utils.Modulo
	httpServer interface{}
)

func main() {
	// Inicializar módulo
	inicializarModulo()

	// Bloquear ejecución para mantener el programa corriendo
	select {}
}

func inicializarModulo() {

	rutaConfig := filepath.Join("configs", "memoria-config.json")

	// Crear módulo
	modulo = utils.NuevoModulo("Memoria", rutaConfig)

	// Inicializar logger
	utils.InicializarLogger("INFO", "Memoria") // Nivel provisional hasta cargar config

	// Cargar configuración
	config = utils.CargarConfiguracion[MemoryConfig](rutaConfig)

	// Actualizar nivel de log con el de la configuración
	utils.InicializarLogger(config.LogLevel, "Memoria")

	// Verificar directorio de dumps
	if err := os.MkdirAll(config.DumpPath, 0755); err != nil {
		utils.InfoLog.Warn("No se pudo crear directorio para dumps", "error", err)
	} else {
		utils.InfoLog.Info("Directorio para dumps verificado", "ruta", config.DumpPath)
	}

	// Inicializar memoria
	inicializarMemoria()

	// Inicializar métricas
	inicializarMetricas()

	// Inicializar mapa de instrucciones por proceso
	instruccionesPorProceso = make(map[int][]string)

	// Registrar handlers en el módulo
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeHandshake), "handshake", handlerHandshake)
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeOperacion), "default", handlerOperacion)
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeObtenerInstruccion), "default", handlerObtenerInstruccion)
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeFetch), "default", handlerObtenerInstruccion)
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeEspacioLibre), "default", handlerEspacioLibre)
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeInicializarProceso), "default", handlerInicializarProceso)
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeFinalizarProceso), "default", handlerFinalizarProceso)
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeLeer), "default", handlerLeerMemoria)
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeEscribir), "default", handlerEscribirMemoria)
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeObtenerMarco), "default", handlerObtenerMarco)

	// Registrar handlers de SWAP
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeSuspenderProceso), "default", handlerSuspenderProceso)
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeDessuspenderProceso), "default", handlerDessuspenderProceso)

	// Registrar handler de Memory Dump
	modulo.RegistrarHandler(strconv.Itoa(utils.MensajeMemoryDump), "default", handlerMemoryDump)

	// Iniciar servidor
	modulo.IniciarServidor(config.IPMemory, config.PortMemory)

	// Obtener el servidor para iniciar en la goroutine principal
	httpServer = modulo.Server
}

// Handlers
func handlerHandshake(msg *utils.Mensaje) (interface{}, error) {
	utils.InfoLog.Info("Handshake recibido", "origen", msg.Origen)

	// Aplicar retardo de memoria
	utils.AplicarRetardo("handshake", config.MemoryDelay)

	return map[string]interface{}{
		"status":           "OK",
		"tam_pagina":       config.PageSize,
		"entradas_por_pag": config.EntriesPerPage,
		"niveles":          config.NumberOfLevels,
	}, nil
}

func procesarOperacion(msg *utils.Mensaje) (interface{}, error) {
	tipoOperacion := utils.ObtenerTipoOperacion(msg, "memoria")
	utils.InfoLog.Info("Operación procesada", "tipo", tipoOperacion)

	return map[string]interface{}{
		"status":  "OK",
		"mensaje": "Operación de memoria completada exitosamente",
	}, nil
}
