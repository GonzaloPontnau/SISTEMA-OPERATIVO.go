package main

import (
	"fmt"
	"path/filepath"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

func handlerOperacion(msg *utils.Mensaje) (interface{}, error) {
	// Determinar el tipo de operación
	tipoOperacion := utils.ObtenerTipoOperacion(msg, "memoria")

	// Seleccionar el retardo adecuado
	retardo := config.MemoryDelay
	if tipoOperacion == "swap" {
		retardo = config.SwapDelay
	}

	return utils.HandlerGenerico(msg, retardo, procesarOperacion)
}

func handlerObtenerInstruccion(msg *utils.Mensaje) (interface{}, error) {
	// Extraer el PID del mensaje
	datos := msg.Datos.(map[string]interface{})
	pid, ok := datos["pid"].(float64)
	if !ok {
		return map[string]interface{}{
			"error": "PID no proporcionado o formato incorrecto",
		}, nil
	}

	// Convertir PID a entero
	pidInt := int(pid)

	// Verificar si hay instrucciones para el PID
	instrucciones, existe := instruccionesPorProceso[pidInt]
	if !existe || len(instrucciones) == 0 {
		// Intentar cargar las instrucciones si no existen
		if err := cargarInstrucciones(pidInt); err != nil {
			return map[string]interface{}{
				"error": fmt.Sprintf("No se pudieron cargar instrucciones para el PID %d: %v", pidInt, err),
			}, nil
		}
		instrucciones = instruccionesPorProceso[pidInt]
	}

	// Obtener la próxima instrucción
	instruccion := instrucciones[0]

	// Actualizar la lista de instrucciones (eliminar la primera)
	instruccionesPorProceso[pidInt] = instrucciones[1:]

	// Log de la instrucción obtenida
	utils.InfoLog.Info(fmt.Sprintf("## PID: %d - Obtener instrucción: %d - Instrucción: %s", pidInt, len(instruccionesPorProceso[pidInt]), instruccion))

	// Responder con la instrucción
	return map[string]interface{}{
		"status":      "OK",
		"instruccion": instruccion,
	}, nil
}

func handlerEspacioLibre(msg *utils.Mensaje) (interface{}, error) {
	// Calcular espacio libre real
	espacioLibre := calcularEspacioLibre()

	// Log del espacio libre solicitado
	utils.InfoLog.Info("Espacio libre solicitado", "espacio_libre", espacioLibre)

	// Responder con el espacio libre
	return map[string]interface{}{
		"status":        "OK",
		"espacio_libre": espacioLibre,
	}, nil
}

// calcularEspacioLibre calcula el espacio libre total en bytes
// basado en los marcos libres disponibles
func calcularEspacioLibre() int {
	espacioLibre := 0
	for _, libre := range marcosLibres {
		if libre {
			espacioLibre += config.PageSize
		}
	}
	return espacioLibre
}

func handlerInicializarProceso(msg *utils.Mensaje) (interface{}, error) {
	datos, ok := msg.Datos.(map[string]interface{})
	if !ok {
		return map[string]interface{}{"error": "Formato de datos incorrecto"}, nil
	}

	pid := int(datos["pid"].(float64))
	tamanio := int(datos["tamanio"].(float64))
	archivoOrigen := datos["archivo"].(string)

	// Verificar si hay suficiente espacio libre
	if calcularEspacioLibre() < tamanio {
		return map[string]interface{}{
			"error": fmt.Sprintf("No hay suficiente espacio libre para inicializar el proceso %d", pid),
		}, nil
	}

	// Copiar el archivo de pseudocódigo
	destino := filepath.Join(config.ScriptsPath, fmt.Sprintf("%d.txt", pid))
	if err := copiarPseudocodigo(archivoOrigen, destino); err != nil {
		utils.ErrorLog.Error("Error copiando pseudocódigo", "error", err)
		return map[string]interface{}{"error": err.Error()}, nil
	}

	// Crear tablas de páginas para el proceso
	_, err := crearTablasPaginas(pid, tamanio)
	if err != nil {
		utils.ErrorLog.Error("Error creando tablas de páginas", "error", err)
		return map[string]interface{}{"error": err.Error()}, nil
	}

	// Cargar instrucciones en memoria para ese PID
	if err := cargarInstrucciones(pid); err != nil {
		utils.ErrorLog.Error("Error cargando instrucciones", "error", err)
		// Liberar memoria asignada en caso de error
		liberarMemoriaProceso(pid)
		return map[string]interface{}{"error": err.Error()}, nil
	}

	utils.InfoLog.Info("Proceso inicializado correctamente",
		"PID", pid,
		"tamaño", tamanio,
		"archivo", archivoOrigen)

	return map[string]interface{}{
		"status": "OK",
	}, nil
}

func handlerFinalizarProceso(msg *utils.Mensaje) (interface{}, error) {
	// Extraer el PID del mensaje
	datos := msg.Datos.(map[string]interface{})
	pid, ok := datos["pid"].(float64)
	if !ok {
		return map[string]interface{}{
			"error": "PID no proporcionado o formato incorrecto",
		}, nil
	}

	// Convertir PID a entero
	pidInt := int(pid)

	// Liberar memoria del proceso
	if err := liberarMemoriaProceso(pidInt); err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Error al liberar memoria del proceso %d: %v", pidInt, err),
		}, nil
	}

	// Log de finalización con métricas
	if metricas, existe := metricasPorProceso[pidInt]; existe {
		utils.InfoLog.Info(fmt.Sprintf("## PID: %d Proceso Destruido Métricas Acc.T.Pag: %d; Inst.Sol.: %d; SWAP: %d; Mem. Prin.: %d; Lec.Mem.: %d; Esc.Mem.: %d",
			pidInt,
			metricas.AccesosTablasPaginas,
			metricas.InstruccionesSolicitadas,
			metricas.BajadasSwap,
			metricas.SubidasMemoria,
			metricas.LecturasMemoria,
			metricas.EscriturasMemoria))

		// Eliminar las métricas del proceso finalizado
		delete(metricasPorProceso, pidInt)
	}

	// Eliminar instrucciones del proceso
	delete(instruccionesPorProceso, pidInt)

	return map[string]interface{}{
		"status": "OK",
	}, nil
}

func handlerLeerMemoria(msg *utils.Mensaje) (interface{}, error) {
	// Extraer datos del mensaje
	datos := msg.Datos.(map[string]interface{})
	pid, ok := datos["pid"].(float64)
	if !ok {
		return map[string]interface{}{"error": "PID no proporcionado o formato incorrecto"}, nil
	}
	pidInt := int(pid)

	// Dirección y tamaño pueden ser físicos o lógicos según lo que necesitemos
	dirFisica, ok := datos["direccion_fisica"].(float64)
	if !ok {
		// Si no se proporcionó dirección física, intentar con dirección lógica
		dirLogica, ok := datos["direccion_logica"].(float64)
		if !ok {
			return map[string]interface{}{"error": "Dirección no proporcionada o formato incorrecto"}, nil
		}

		// Traducir dirección lógica a física
		dirFisicaInt, err := traducirDireccion(pidInt, int(dirLogica))
		if err != nil {
			return map[string]interface{}{"error": fmt.Sprintf("Error traduciendo dirección: %v", err)}, nil
		}
		dirFisica = float64(dirFisicaInt)
	}

	tamanio, ok := datos["tamanio"].(float64)
	if !ok {
		tamanio = 1 // Por defecto, leer un byte
	}

	// Verificar límites
	if int(dirFisica) < 0 || int(dirFisica)+int(tamanio) > len(memoriaPrincipal) {
		return map[string]interface{}{"error": "Dirección fuera de rango"}, nil
	}

	// Leer de memoria
	valor := memoriaPrincipal[int(dirFisica):int(dirFisica+tamanio)]

	// Actualizar métricas
	actualizarMetricasLectura(pidInt)

	// Log obligatorio
	utils.InfoLog.Info(fmt.Sprintf("## PID: %d Lectura Dir. Física: %d Tamaño: %d",
		pidInt, int(dirFisica), int(tamanio)))

	// Responder con el valor leído
	return map[string]interface{}{
		"status": "OK",
		"valor":  string(valor), // Convertir a string para respuesta JSON
	}, nil
}

func handlerEscribirMemoria(msg *utils.Mensaje) (interface{}, error) {
	// Extraer datos del mensaje
	datos := msg.Datos.(map[string]interface{})
	pid, ok := datos["pid"].(float64)
	if !ok {
		return map[string]interface{}{"error": "PID no proporcionado o formato incorrecto"}, nil
	}
	pidInt := int(pid)

	// Dirección puede ser física o lógica según lo que necesitemos
	dirFisica, ok := datos["direccion_fisica"].(float64)
	if !ok {
		// Si no se proporcionó dirección física, intentar con dirección lógica
		dirLogica, ok := datos["direccion_logica"].(float64)
		if !ok {
			return map[string]interface{}{"error": "Dirección no proporcionada o formato incorrecto"}, nil
		}

		// Traducir dirección lógica a física
		dirFisicaInt, err := traducirDireccion(pidInt, int(dirLogica))
		if err != nil {
			return map[string]interface{}{"error": fmt.Sprintf("Error traduciendo dirección: %v", err)}, nil
		}
		dirFisica = float64(dirFisicaInt)
	}

	valor, ok := datos["valor"].(string)
	if !ok {
		return map[string]interface{}{"error": "Valor no proporcionado o formato incorrecto"}, nil
	}

	// Verificar límites
	if int(dirFisica) < 0 || int(dirFisica)+len(valor) > len(memoriaPrincipal) {
		return map[string]interface{}{"error": "Dirección fuera de rango"}, nil
	}

	// Escribir en memoria
	copy(memoriaPrincipal[int(dirFisica):int(dirFisica)+len(valor)], []byte(valor))

	// Actualizar métricas
	actualizarMetricasEscritura(pidInt)

	// Log obligatorio
	utils.InfoLog.Info(fmt.Sprintf("## PID: %d Escritura Dir. Física: %d Tamaño: %d",
		pidInt, int(dirFisica), len(valor)))

	// Responder OK
	return map[string]interface{}{
		"status": "OK",
	}, nil
}

func handlerObtenerMarco(msg *utils.Mensaje) (interface{}, error) {
	// Extraer datos del mensaje
	datos := msg.Datos.(map[string]interface{})
	pid, ok := datos["pid"].(float64)
	if !ok {
		return map[string]interface{}{"error": "PID no proporcionado o formato incorrecto"}, nil
	}
	pidInt := int(pid)

	numPagina, ok := datos["pagina"].(float64)
	if !ok {
		return map[string]interface{}{"error": "Número de página no proporcionado o formato incorrecto"}, nil
	}

	// Obtener la tabla de páginas del proceso
	tabla, existe := tablasPaginas[pidInt]
	if !existe {
		return map[string]interface{}{"error": "No existe tabla de páginas para el PID proporcionado"}, nil
	}

	// Obtener el marco para la página solicitada
	marco, err := obtenerMarcoDesdeTabla(pidInt, tabla, int(numPagina), 1)
	if err != nil {
		return map[string]interface{}{"error": fmt.Sprintf("Error obteniendo marco: %v", err)}, nil
	}

	// Log obligatorio
	utils.InfoLog.Info(fmt.Sprintf("PID: %d OBTENER MARCO Página: %d Marco: %d",
		pidInt, int(numPagina), marco))

	// Responder con el marco
	return map[string]interface{}{
		"status": "OK",
		"marco":  marco,
	}, nil
}

// handlerSuspenderProceso atiende peticiones de suspensión del Kernel
func handlerSuspenderProceso(msg *utils.Mensaje) (interface{}, error) {
	// Extraer el PID del mensaje
	datos := msg.Datos.(map[string]interface{})
	pid, ok := datos["pid"].(float64)
	if !ok {
		return map[string]interface{}{
			"error": "PID no proporcionado o formato incorrecto",
		}, nil
	}
	pidInt := int(pid)

	// Suspender el proceso
	err := suspenderProceso(pidInt)
	if err != nil {
		utils.ErrorLog.Error("Error al suspender proceso", "pid", pidInt, "error", err)
		return map[string]interface{}{
			"error": err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"status": "OK",
	}, nil
}

// handlerDessuspenderProceso atiende peticiones de dessuspensión del Kernel
func handlerDessuspenderProceso(msg *utils.Mensaje) (interface{}, error) {
	// Extraer el PID del mensaje
	datos := msg.Datos.(map[string]interface{})
	pid, ok := datos["pid"].(float64)
	if !ok {
		return map[string]interface{}{
			"error": "PID no proporcionado o formato incorrecto",
		}, nil
	}
	pidInt := int(pid)

	// Dessuspender el proceso
	err := dessuspenderProceso(pidInt)
	if err != nil {
		utils.ErrorLog.Error("Error al dessuspender proceso", "pid", pidInt, "error", err)
		return map[string]interface{}{
			"error": err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"status": "OK",
	}, nil
}