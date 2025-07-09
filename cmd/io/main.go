package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	// Verificar si se proporcionó el nombre del dispositivo IO
	if len(os.Args) < 2 {
		fmt.Println("Uso: ./io <nombre_dispositivo> [-c ruta_configuracion]")
		os.Exit(1)
	}

	nombreDispositivo := os.Args[1]

	// Path de configuración por defecto
	rutaConfig := filepath.Join("configs", "io-config.json")

	// Verificar si se especificó una ruta de configuración personalizada
	for i := 2; i < len(os.Args); i++ {
		if (os.Args[i] == "-c" || os.Args[i] == "--config") && i+1 < len(os.Args) {
			rutaConfig = os.Args[i+1]
			break
		}
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

	utils.InfoLog.Info("Módulo IO inicializado", "nombre", nombreDispositivo)

	// Registrar handlers
	modulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeHandshake), "handshake", handlerHandshake)
	modulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeOperacion), "EJECUTAR_PROCESO", handlerOperacion)
	modulo.RegistrarHandler(fmt.Sprintf("%d", utils.MensajeEjecutar), "default", handlerOperacion)

	// Iniciar servidor
	modulo.IniciarServidorAutoPuerto(config.IPIO)

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
