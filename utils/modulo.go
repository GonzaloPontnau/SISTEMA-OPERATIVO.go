package utils

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Modulo representa un módulo genérico del sistema
type Modulo struct {
	Nombre      string
	Server      *HTTPServer
	Clientes    map[string]*HTTPClient
	ConfigPath  string
	HandlerFunc map[string]map[string]HTTPHandlerFunc // Mapa de tipo -> operación -> handler
}

// NuevoModulo crea una nueva instancia de un módulo
func NuevoModulo(nombre string, configPath string) *Modulo {
	return &Modulo{
		Nombre:      nombre,
		Clientes:    make(map[string]*HTTPClient),
		ConfigPath:  configPath,
		HandlerFunc: make(map[string]map[string]HTTPHandlerFunc),
	}
}

// RegistrarHandler registra un handler para un tipo de mensaje y operación específicos
func (m *Modulo) RegistrarHandler(tipo string, operacion string, handler HTTPHandlerFunc) {
	// Asegurar que existe el mapa para este tipo de mensaje
	if _, existe := m.HandlerFunc[tipo]; !existe {
		m.HandlerFunc[tipo] = make(map[string]HTTPHandlerFunc)
	}
	// Registrar el handler para esta operación
	m.HandlerFunc[tipo][operacion] = handler
	slog.Debug("Handler registrado", "tipo", tipo, "operacion", operacion)
}

// IniciarServidor crea e inicializa el servidor HTTP del módulo
func (m *Modulo) IniciarServidor(ip string, puerto int) {
	// Crear el servidor HTTP
	m.Server = NewHTTPServer(ip, puerto, m.Nombre)

	// Registrar handlers para el servidor HTTP
	for tipoStr, handlersPorOperacion := range m.HandlerFunc {
		tipo, err := strconv.Atoi(tipoStr)
		if err != nil {
			slog.Error("Error al convertir tipo de mensaje a entero", "tipo", tipoStr, "error", err)
			continue
		}

		// Registrar handler para este tipo de mensaje en el servidor HTTP
		m.Server.RegisterHTTPHandler(tipo, func(msg *Mensaje) (interface{}, error) {
			// Buscar handler específico para esta operación
			operacion := msg.Operacion
			if operacion == "" {
				operacion = "default" // Usar default si no se especifica operación
			}

			handler, existe := handlersPorOperacion[operacion]
			if !existe {
				handler, existe = handlersPorOperacion["default"]
				if !existe {
					slog.Error("No hay handler para operación", "tipo", tipo, "operacion", operacion)
					return nil, fmt.Errorf("no hay handler para operación %s", operacion)
				}
			}

			// Ejecutar el handler
			return handler(msg)
		})
	}

	// Iniciar el servidor en una goroutine
	go func() {
		err := m.Server.Start()
		if err != nil {
			slog.Error("Error al iniciar servidor HTTP", "error", err)
			os.Exit(1)
		}
	}()

	// Log de registro de handlers
	slog.Info("Servidor HTTP iniciado", "módulo", m.Nombre, "ip", ip, "puerto", puerto)
}

// CrearCliente crea un cliente HTTP para conectarse a otro módulo
func (m *Modulo) CrearCliente(nombre string, ip string, puerto int) {
	m.Clientes[nombre] = NewHTTPClient(ip, puerto, m.Nombre)
	slog.Info("Cliente HTTP creado", "módulo_origen", m.Nombre, "módulo_destino", nombre)
}

// ConectarCliente intenta conectar con un módulo con reintentos indefinidos
func (m *Modulo) ConectarCliente(nombreModulo string, tiempoReintento int, datosHandshake map[string]interface{}) {
	cliente, existe := m.Clientes[nombreModulo]
	if !existe {
		slog.Error("Cliente no encontrado, no se puede conectar", "módulo_destino", nombreModulo)
		return
	}

	slog.Info("Intentando conectar con módulo (reintentos indefinidos)", "destino", nombreModulo)

	for i := 1; ; i++ {
		_, err := cliente.EnviarHTTPMensaje(MensajeHandshake, "handshake", datosHandshake)
		if err == nil {
			slog.Info("Conexión establecida exitosamente", "destino", nombreModulo)
			return // Salir del bucle y de la función
		}

		slog.Warn("Error al conectar, reintentando...",
			"destino", nombreModulo,
			"intento", i,
			"error", err,
			"proximo_intento_en", tiempoReintento)
		time.Sleep(time.Duration(tiempoReintento) * time.Second)
	}
}

// EnviarMensaje envía un mensaje a un módulo destino específico
func (m *Modulo) EnviarMensaje(nombreModulo string, tipoMensaje int, operacion string, datos map[string]interface{}) (interface{}, error) {
	// Obtener el cliente correspondiente
	cliente, existe := m.Clientes[nombreModulo]
	if !existe {
		return nil, fmt.Errorf("cliente no encontrado para el módulo %s", nombreModulo)
	}

	// Enviar el mensaje y retornar la respuesta
	response, err := cliente.EnviarHTTPMensaje(tipoMensaje, operacion, datos)
	if err != nil {
		return nil, err
	}

	// Devolver la respuesta sin hacer type assertion, para que el llamador haga la conversión necesaria
	return response, nil
}

// CargarConfiguracion carga la configuración del módulo
func CargarConfiguracion[T any](ruta string) *T {
	slog.Info("Cargando configuración", "ruta", ruta)

	// Asegurarse que el directorio existe
	dir := filepath.Dir(ruta)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("Error al crear directorio de configuración", "error", err)
		os.Exit(1)
	}

	config, err := LoadConfig[T](ruta)
	if err != nil {
		slog.Error("Error cargando configuración", "error", err)
		os.Exit(1)
	}

	slog.Info("Configuración cargada correctamente")
	return config
}

// Constantes para tipos de mensajes entre módulos
const (
	// Mensaje para solicitar lectura de memoria
	MensajeLeer = iota + 1
	// Mensaje para solicitar escritura en memoria
	MensajeEscribir
	// Mensaje para obtener el número de marco de página
	MensajeObtenerMarco
	MensajeFetch
	MensajeEjecutar
	MensajeObtenerInstruccion
	MensajeEspacioLibre        = 7
	MensajeInicializarProceso  = 42
	MensajeInterrupcion        = 50
	MensajeFinalizarProceso    = 43
	MensajeDessuspenderProceso = 44
	MensajeSuspenderProceso    = 45
	MensajeMemoryDump          = 46
)
