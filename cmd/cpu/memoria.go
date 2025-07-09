package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

// Agregar a memoria.go - Estructura para leer config de memoria
type MemoriaConfig struct {
	PortMemory     int    `json:"port_memory"`
	IPMemory       string `json:"ip_memory"`
	MemorySize     int    `json:"memory_size"`
	PageSize       int    `json:"page_size"`
	EntriesPerPage int    `json:"entries_per_page"`
	NumberOfLevels int    `json:"number_of_levels"`
	MemoryDelay    int    `json:"memory_delay"`
	SwapfilePath   string `json:"swapfile_path"`
	SwapDelay      int    `json:"swap_delay"`
	LogLevel       string `json:"log_level"`
	DumpPath       string `json:"dump_path"`
	ScriptsPath    string `json:"scripts_path"`
}

// Configuración de memoria obtenida dinámicamente
var (
	tamanoPagina     int
	entradasPorTabla int
	numeroDeNiveles  int
	configCargada    bool = false
)

// Cargar configuración directamente desde memoria-config.json
func cargarConfigMemoria() error {
	if configCargada {
		return nil // Ya está cargada
	}

	utils.InfoLog.Info("Cargando configuración de memoria desde archivo")

	rutaConfigMemoria := filepath.Join("configs", "memoria-config.json")

	// Verificar que el archivo existe
	if _, err := os.Stat(rutaConfigMemoria); os.IsNotExist(err) {
		utils.ErrorLog.Error("Archivo de configuración de memoria no encontrado", "ruta", rutaConfigMemoria)
		return fmt.Errorf("archivo de configuración de memoria no encontrado: %s", rutaConfigMemoria)
	}

	// Cargar configuración de memoria
	configMemoria := utils.CargarConfiguracion[MemoriaConfig](rutaConfigMemoria)

	// Extraer valores necesarios
	tamanoPagina = configMemoria.PageSize
	entradasPorTabla = configMemoria.EntriesPerPage
	numeroDeNiveles = configMemoria.NumberOfLevels

	configCargada = true

	utils.InfoLog.Info("Configuración de memoria cargada desde archivo",
		"page_size", tamanoPagina,
		"entries_per_page", entradasPorTabla,
		"number_of_levels", numeroDeNiveles,
		"archivo", rutaConfigMemoria)

	return nil
}

func calcularEntradasNiveles(direccionLogica int) ([]int, int) {
	numeroPagina := direccionLogica / tamanoPagina
	desplazamiento := direccionLogica % tamanoPagina

	entradas := make([]int, numeroDeNiveles)

	for nivel := 0; nivel < numeroDeNiveles; nivel++ {
		exponente := numeroDeNiveles - nivel - 1
		potencia := int(math.Pow(float64(entradasPorTabla), float64(exponente)))
		entradas[nivel] = (numeroPagina / potencia) % entradasPorTabla
	}

	return entradas, desplazamiento
}

// Traducir dirección lógica a física
func traducirDireccion(pid, direccionLogica int) int {
	// Asegurar que la configuración esté cargada
	if !configCargada {
		err := cargarConfigMemoria()
		if err != nil {
			utils.ErrorLog.Error("Error obteniendo configuración", "error", err)
			return -1
		}
	}

	// Calcular número de página y desplazamiento usando config dinámica
	numeroPagina := int(math.Floor(float64(direccionLogica) / float64(tamanoPagina)))
	desplazamiento := direccionLogica % tamanoPagina

	// Buscar en TLB si está habilitada
	if config.TLBEntries > 0 {
		marco := buscarEnTLB(pid, numeroPagina)
		if marco != -1 {
			// TLB Hit
			utils.InfoLog.Info(fmt.Sprintf("PID: %d - TLB HIT - Pagina: %d", pid, numeroPagina))
			return marco*tamanoPagina + desplazamiento
		} else {
			// TLB Miss
			utils.InfoLog.Info(fmt.Sprintf("PID: %d - TLB MISS - Pagina: %d", pid, numeroPagina))
		}
	}

	// Obtener marco desde memoria
	marco := obtenerMarcoDeMemoria(pid, numeroPagina)
	if marco == -1 {
		utils.ErrorLog.Error("Error obteniendo marco de memoria", "pid", pid, "pagina", numeroPagina)
		return -1
	}

	// Actualizar TLB si está habilitada
	if config.TLBEntries > 0 {
		actualizarTLB(pid, numeroPagina, marco)
	}

	return marco*tamanoPagina + desplazamiento
}

// Buscar página en TLB
func buscarEnTLB(pid, numeroPagina int) int {
	mutex.Lock()
	defer mutex.Unlock()

	for i, entrada := range tlbEntries {
		if entrada.PageNumber == numeroPagina && entrada.PID == pid {
			// Actualizar para LRU si es necesario
			if config.TLBReplacement == "LRU" {
				tlbEntries[i].LastUsed = time.Now().UnixNano()
			}
			return entrada.FrameNumber
		}
	}

	return -1 // No encontrado
}

// Actualizar TLB con nueva entrada
func actualizarTLB(pid, numeroPagina, marco int) {
	mutex.Lock()
	defer mutex.Unlock()

	// Buscar entrada libre
	indiceLibre := -1
	for i, entrada := range tlbEntries {
		if entrada.PageNumber == -1 {
			indiceLibre = i
			break
		}
	}

	// Si hay espacio, agregar
	if indiceLibre != -1 {
		tlbEntries[indiceLibre] = TLBEntry{
			PageNumber:  numeroPagina,
			FrameNumber: marco,
			PID:         pid,
			LastUsed:    time.Now().UnixNano(),
			LoadTime:    tlbCounter,
		}
		tlbCounter++
		return
	}

	// No hay espacio, aplicar algoritmo de reemplazo
	indiceVictima := 0

	if config.TLBReplacement == "FIFO" {
		// Buscar la entrada más antigua
		tiempoMasAntiguo := tlbEntries[0].LoadTime
		for i, entrada := range tlbEntries {
			if entrada.LoadTime < tiempoMasAntiguo {
				tiempoMasAntiguo = entrada.LoadTime
				indiceVictima = i
			}
		}
	} else if config.TLBReplacement == "LRU" {
		// Buscar la entrada menos usada recientemente
		menosUsada := tlbEntries[0].LastUsed
		for i, entrada := range tlbEntries {
			if entrada.LastUsed < menosUsada {
				menosUsada = entrada.LastUsed
				indiceVictima = i
			}
		}
	}

	// Reemplazar víctima
	tlbEntries[indiceVictima] = TLBEntry{
		PageNumber:  numeroPagina,
		FrameNumber: marco,
		PID:         pid,
		LastUsed:    time.Now().UnixNano(),
		LoadTime:    tlbCounter,
	}
	tlbCounter++
}

// Obtener marco de memoria para una página
func obtenerMarcoDeMemoria(pid, numeroPagina int) int {
	utils.InfoLog.Info(fmt.Sprintf("PID: %d - OBTENER MARCO - Página: %d", pid, numeroPagina))

	// Verificar en caché si está habilitada
	if config.CacheEntries > 0 {
		if indiceCache := buscarEnCache(pid, numeroPagina); indiceCache != -1 {
			return indiceCache
		}
	}

	// NUEVO: Calcular entradas multinivel
	direccionLogica := numeroPagina * tamanoPagina
	entradas, _ := calcularEntradasNiveles(direccionLogica)

	// ACTUALIZADO: Preparar mensaje con info multinivel
	params := map[string]interface{}{
		"pid":              pid,
		"pagina":           numeroPagina,
		"entradas_niveles": entradas,        // NUEVO
		"niveles":          numeroDeNiveles, // NUEVO
	}

	// Simular delay de cache si está configurado
	if config.CacheDelay > 0 {
		time.Sleep(time.Duration(config.CacheDelay) * time.Millisecond)
	}

	// Enviar solicitud a memoria
	respuesta, err := modulo.EnviarMensaje("Memoria", utils.MensajeObtenerMarco, "OBTENER_MARCO", params)
	if err != nil {
		utils.ErrorLog.Error("Error al solicitar marco a memoria", "error", err)
		return -1
	}

	// Extraer marco de la respuesta
	datos, ok := respuesta.(map[string]interface{})
	if !ok {
		utils.ErrorLog.Error("Formato de datos incorrecto")
		return -1
	}

	// MEJORADO: Manejar marco como float64 o int
	var marcoInt int
	if marco, ok := datos["marco"].(float64); ok {
		marcoInt = int(marco)
	} else if marco, ok := datos["marco"].(int); ok {
		marcoInt = marco
	} else {
		utils.ErrorLog.Error("Formato de marco incorrecto", "marco", datos["marco"])
		return -1
	}

	utils.InfoLog.Info(fmt.Sprintf("PID: %d - OBTENER MARCO - Página: %d - Marco: %d", pid, numeroPagina, marcoInt))

	// Si la caché está habilitada, actualizar
	if config.CacheEntries > 0 {
		actualizarCache(pid, numeroPagina)
	}

	return marcoInt
}

// Buscar en caché
func buscarEnCache(pid, numeroPagina int) int {
	mutex.Lock()
	defer mutex.Unlock()

	for i, entrada := range cacheEntries {
		if entrada.PageNumber == numeroPagina && entrada.PID == pid {
			// Cache Hit
			utils.InfoLog.Info(fmt.Sprintf("PID: %d - Cache Hit - Pagina: %d", pid, numeroPagina))

			// Actualizar bit de referencia para CLOCK
			cacheEntries[i].Referenced = true

			return i // Simplificado: usamos índice como marco
		}
	}

	// Cache Miss
	utils.InfoLog.Info(fmt.Sprintf("PID: %d - Cache Miss - Pagina: %d", pid, numeroPagina))
	return -1
}

// Actualizar caché
func actualizarCache(pid, numeroPagina int) {
	mutex.Lock()
	defer mutex.Unlock()

	// Buscar entrada libre
	indiceLibre := -1
	for i, entrada := range cacheEntries {
		if entrada.PageNumber == -1 {
			indiceLibre = i
			break
		}
	}

	// Si hay espacio, agregar
	if indiceLibre != -1 {
		cacheEntries[indiceLibre] = CacheEntry{
			PageNumber: numeroPagina,
			Content:    "", // Contenido se obtendría al leer
			PID:        pid,
			Modified:   false,
			Referenced: true,
		}
		utils.InfoLog.Info(fmt.Sprintf("PID: %d - Cache Add - Pagina: %d", pid, numeroPagina))
		return
	}

	// No hay espacio, aplicar algoritmo de reemplazo
	if config.CacheReplacement == "CLOCK" {
		aplicarCLOCK(pid, numeroPagina)
	} else if config.CacheReplacement == "CLOCK-M" {
		aplicarCLOCKM(pid, numeroPagina)
	}
}

// Algoritmo CLOCK para caché
func aplicarCLOCK(pid, numeroPagina int) {
	for {
		// Si la entrada actual no está referenciada, es la víctima
		if !cacheEntries[clockPointer].Referenced {
			// Si está modificada, actualizar en memoria
			if cacheEntries[clockPointer].Modified {
				actualizarMemoria(cacheEntries[clockPointer].PID, cacheEntries[clockPointer].PageNumber)
			}

			// Reemplazar
			cacheEntries[clockPointer] = CacheEntry{
				PageNumber: numeroPagina,
				Content:    "", // Contenido se obtendría al leer
				PID:        pid,
				Modified:   false,
				Referenced: true,
			}
			utils.InfoLog.Info(fmt.Sprintf("PID: %d - Cache Add - Pagina: %d", pid, numeroPagina))

			// Avanzar puntero
			clockPointer = (clockPointer + 1) % len(cacheEntries)
			return
		}

		// Si está referenciada, quitar referencia y seguir
		cacheEntries[clockPointer].Referenced = false
		clockPointer = (clockPointer + 1) % len(cacheEntries)
	}
}

// Algoritmo CLOCK-M para caché
func aplicarCLOCKM(pid, numeroPagina int) {
	// Primera vuelta: buscar (0,0) - no referenciada, no modificada
	punteroInicial := clockPointer
	for {
		if !cacheEntries[clockPointer].Referenced && !cacheEntries[clockPointer].Modified {
			// Encontrada víctima perfecta
			break
		}
		clockPointer = (clockPointer + 1) % len(cacheEntries)
		if clockPointer == punteroInicial {
			break // Completó la vuelta
		}
	}

	// Si no encontró (0,0), segunda vuelta: buscar (0,1) - no referenciada, modificada
	if clockPointer == punteroInicial {
		for {
			if !cacheEntries[clockPointer].Referenced && cacheEntries[clockPointer].Modified {
				break
			}
			clockPointer = (clockPointer + 1) % len(cacheEntries)
			if clockPointer == punteroInicial {
				break
			}
		}
	}

	// Si no encontró (0,1), tercera vuelta: quitar referencias y volver a empezar
	if clockPointer == punteroInicial {
		for i := range cacheEntries {
			cacheEntries[i].Referenced = false
		}

		// Ahora buscar no referenciada
		for {
			if !cacheEntries[clockPointer].Referenced {
				break
			}
			clockPointer = (clockPointer + 1) % len(cacheEntries)
		}
	}

	// Si está modificada, actualizar en memoria
	if cacheEntries[clockPointer].Modified {
		actualizarMemoria(cacheEntries[clockPointer].PID, cacheEntries[clockPointer].PageNumber)
	}

	// Reemplazar
	cacheEntries[clockPointer] = CacheEntry{
		PageNumber: numeroPagina,
		Content:    "", // Contenido se obtendría al leer
		PID:        pid,
		Modified:   false,
		Referenced: true,
	}
	utils.InfoLog.Info(fmt.Sprintf("PID: %d - Cache Add - Pagina: %d", pid, numeroPagina))

	// Avanzar puntero
	clockPointer = (clockPointer + 1) % len(cacheEntries)
}

// Actualizar memoria desde caché
func actualizarMemoria(pid, numeroPagina int) {
	// Buscar entrada en caché
	var contenido string
	for _, e := range cacheEntries {
		if e.PageNumber == numeroPagina && e.PID == pid {
			contenido = e.Content
			break
		}
	}

	// Obtener marco desde memoria
	marco := obtenerMarcoDeMemoria(pid, numeroPagina)

	if marco != -1 {
		// Preparar mensaje para memoria
		params := map[string]interface{}{
			"pid":       pid,
			"marco":     marco,
			"contenido": contenido,
		}

		// Enviar solicitud a memoria
		_, err := modulo.EnviarMensaje("Memoria", utils.MensajeEscribir, "ESCRIBIR", params)
		if err != nil {
			utils.ErrorLog.Error("Error al actualizar memoria", "error", err)
			return
		}

		utils.InfoLog.Info(fmt.Sprintf("PID: %d - Memory Update - Página: %d - Frame: %d", pid, numeroPagina, marco))
	}
}

// Escribir en memoria
func escribirEnMemoria(pid, direccionLogica int, valor string) {
	direccionFisica := traducirDireccion(pid, direccionLogica)

	// Verificar si está en caché
	if config.CacheEntries > 0 {
		numeroPagina := int(math.Floor(float64(direccionLogica) / float64(tamanoPagina)))

		mutex.Lock()
		for i, entrada := range cacheEntries {
			if entrada.PageNumber == numeroPagina && entrada.PID == pid {
				// Actualizar en caché
				cacheEntries[i].Content = valor
				cacheEntries[i].Modified = true
				cacheEntries[i].Referenced = true

				utils.InfoLog.Info(fmt.Sprintf("PID: %d - Acción: ESCRIBIR - Dirección Física: %d - Valor: %s", pid, direccionFisica, valor))
				mutex.Unlock()
				return
			}
		}
		mutex.Unlock()
	}

	// No está en caché o caché deshabilitada, escribir en memoria
	params := map[string]interface{}{
		"pid":       pid,
		"direccion": direccionFisica,
		"valor":     valor,
	}

	// Enviar solicitud a memoria
	_, err := modulo.EnviarMensaje("Memoria", utils.MensajeEscribir, "ESCRIBIR", params)
	if err != nil {
		utils.ErrorLog.Error("Error al escribir en memoria", "error", err)
		return
	}

	utils.InfoLog.Info(fmt.Sprintf("PID: %d - Acción: ESCRIBIR - Dirección Física: %d - Valor: %s", pid, direccionFisica, valor))
}

// Leer de memoria
func leerDeMemoria(pid, direccionLogica, tamano int) string {
	direccionFisica := traducirDireccion(pid, direccionLogica)

	// Verificar si está en caché
	if config.CacheEntries > 0 {
		tamanoPagina := 256 // Ejemplo
		numeroPagina := int(math.Floor(float64(direccionLogica) / float64(tamanoPagina)))

		mutex.Lock()
		for i, entrada := range cacheEntries {
			if entrada.PageNumber == numeroPagina && entrada.PID == pid {
				// Leer de caché
				valor := entrada.Content
				cacheEntries[i].Referenced = true

				utils.InfoLog.Info(fmt.Sprintf("PID: %d - Acción: LEER - Dirección Física: %d - Valor: %s", pid, direccionFisica, valor))
				mutex.Unlock()
				return valor
			}
		}
		mutex.Unlock()
	}

	// No está en caché o caché deshabilitada, leer de memoria
	params := map[string]interface{}{
		"pid":       pid,
		"direccion": direccionFisica,
		"tamano":    tamano,
	}

	// Enviar solicitud a memoria
	respuesta, err := modulo.EnviarMensaje("Memoria", utils.MensajeLeer, "LEER", params)
	if err != nil {
		utils.ErrorLog.Error("Error al leer de memoria", "error", err)
		return ""
	}

	// Extraer valor de la respuesta
	datos, ok := respuesta.(map[string]interface{})
	if !ok {
		utils.ErrorLog.Error("Formato de datos incorrecto")
		return ""
	}

	valor, ok := datos["valor"].(string)
	if !ok {
		utils.ErrorLog.Error("Formato de instrucción incorrecto")
		return ""
	}

	utils.InfoLog.Info(fmt.Sprintf("PID: %d - Acción: LEER - Dirección Física: %d - Valor: %s", pid, direccionFisica, valor))
	return valor
}
