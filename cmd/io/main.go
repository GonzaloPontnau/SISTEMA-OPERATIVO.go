package main

import (
	"fmt"
	"os"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

// Config representa la configuración específica del módulo IO
type IOConfig struct {
	IPIO        string `json:"IP_IO"`
	PortIO      int    `json:"PUERTO_IO"`
	IPKernel    string `json:"IP_KERNEL"`
	PortKernel  int    `json:"PUERTO_KERNEL"`
	LogLevel    string `json:"LOG_LEVEL"`
	RetardoBase int    `json:"RETARDO_BASE"` // Retardo base para operaciones
}

var (
	modulo *utils.Modulo
	config *IOConfig
)

func main() {
	// Verificar argumentos mínimos
	if len(os.Args) < 3 {
		fmt.Println("Uso: ./io <nombre_dispositivo> <ruta_configuracion>")
		fmt.Println("Ejemplo: ./io DISCO configs/io1-config.json")
		os.Exit(1)
	}

	nombreDispositivo := os.Args[1]
	rutaConfig := os.Args[2]

	// Verificar que el archivo de configuración existe
	if _, err := os.Stat(rutaConfig); os.IsNotExist(err) {
		fmt.Printf("Error: El archivo de configuración '%s' no existe\n", rutaConfig)
		os.Exit(1)
	}

	// Inicializar módulo
	inicializarModulo(rutaConfig, nombreDispositivo)

	// Bloquear ejecución
	select {}
}

func inicializarModulo(rutaConfig string, nombreDispositivo string) {
	// Crear módulo
	modulo = utils.NuevoModulo("IO", rutaConfig)

	// Inicializar logger
	utils.InicializarLogger("INFO", "IO") // Nivel provisional hasta cargar config

	// Cargar configuración
	config = utils.CargarConfiguracion[IOConfig](rutaConfig)

	// Actualizar nivel de log con el de la configuración
	utils.InicializarLogger(config.LogLevel, "IO")

	utils.InfoLog.Info("Módulo IO inicializado",
		"nombre", nombreDispositivo,
		"config", rutaConfig,
		"ip", config.IPIO,
		"puerto", config.PortIO)

	// Registrar handlers
	modulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeHandshake), "handshake", handlerHandshake)
	modulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeOperacion), "EJECUTAR_PROCESO", handlerOperacion)
	modulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeOperacion), "IO_REQUEST", handlerOperacion)
	modulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeEjecutar), "default", handlerOperacion)

	// Iniciar servidor
	modulo.IniciarServidor(config.IPIO, config.PortIO)

	// Crear y conectar cliente a Kernel
	modulo.CrearCliente("Kernel", config.IPKernel, config.PortKernel)

	// Intentar conectar con el Kernel de forma asíncrona
	datosHandshake := map[string]interface{}{
		"nombre": nombreDispositivo,
		"tipo":   "IO" + nombreDispositivo,
		"ip":     config.IPIO,
		"puerto": config.PortIO,
	}
	go modulo.ConectarCliente("Kernel", 2, datosHandshake)
}
