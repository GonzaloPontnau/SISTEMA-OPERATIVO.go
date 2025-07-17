package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

func cargarInstrucciones(pid int) error {
	// Log de creación de proceso
	utils.InfoLog.Info("Creación de proceso detectada", "PID", pid)

	// Construir la ruta del archivo de pseudocódigo
	rutaArchivo := filepath.Clean(filepath.Join(config.ScriptsPath, fmt.Sprintf("%d.txt", pid)))

	// Leer el archivo línea por línea
	contenido, err := os.ReadFile(rutaArchivo)
	if err != nil {
		return fmt.Errorf("error al leer el archivo de pseudocódigo para PID %d: %v", pid, err)
	}

	// Dividir el contenido en líneas (instrucciones)
	instrucciones := strings.Split(string(contenido), "\n")

	// Filtrar líneas vacías
	instruccionesFiltradas := []string{}
	for _, instruccion := range instrucciones {
		if strings.TrimSpace(instruccion) != "" {
			instruccionesFiltradas = append(instruccionesFiltradas, instruccion)
		}
	}

	// Guardar las instrucciones en el mapa
	instruccionesPorProceso[pid] = instruccionesFiltradas

	// Log de creación de proceso
	utils.InfoLog.Info("Proceso Creado",
		"PID", pid,
		"Tamaño", len(instruccionesFiltradas))
	return nil
}

func copiarPseudocodigo(origen string, destino string) error {
	// Si el origen no incluye la ruta scripts/, agregarla
	rutaCompleta := origen
	if !strings.Contains(origen, string(filepath.Separator)) && !strings.HasPrefix(origen, "scripts") {
		rutaCompleta = filepath.Join("scripts", origen)
	}

	input, err := os.ReadFile(rutaCompleta)
	if err != nil {
		return err
	}

	err = os.WriteFile(destino, input, 0644)
	if err != nil {
		return err
	}

	utils.InfoLog.Info("Archivo de pseudocódigo copiado", "origen", rutaCompleta, "destino", destino)
	return nil
}

// suspenderProceso guarda todas las páginas de un proceso en SWAP y libera sus marcos
func suspenderProceso(pid int) error {
	// Obtener marcos asignados al proceso
	marcos, existe := marcosAsignadosPorProceso[pid]
	if !existe {
		return fmt.Errorf("el proceso %d no tiene marcos asignados", pid)
	}

	// Obtener tabla de páginas del proceso
	tabla, existeTabla := tablasPaginas[pid]
	if !existeTabla {
		return fmt.Errorf("el proceso %d no tiene tabla de páginas", pid)
	}

	if err := crearMemoryDump(pid); err != nil {
		utils.ErrorLog.Error("Error al crear dump antes de SWAP", "pid", pid, "error", err)
	}

	// Para cada marco, mover su contenido a SWAP
	for _, marco := range marcos {
		// Buscar la página asociada a este marco
		numPagina := encontrarPaginaPorMarco(pid, tabla, marco, 1)
		if numPagina == -1 {
			utils.InfoLog.Warn("No se encontró página asociada al marco",
				"pid", pid, "marco", marco)
			continue
		}

		// Mover a SWAP
		_, err := moverASwap(pid, numPagina, marco)
		if err != nil {
			utils.ErrorLog.Error("Error moviendo página a SWAP",
				"pid", pid, "pagina", numPagina, "marco", marco, "error", err)
			// Continuamos con las demás páginas
		}

		// Marcar página como no presente
		marcarPaginaNoPresente(pid, tabla, numPagina, 1)

		// Marcar marco como libre
		marcosLibres[marco] = true
	}

	// Liberar la lista de marcos asignados al proceso (pero mantener la tabla)
	marcosAsignadosPorProceso[pid] = []int{}

	utils.InfoLog.Info(fmt.Sprintf("## PID: %d - Proceso suspendido a SWAP", pid))

	return nil
}

// dessuspenderProceso carga todas las páginas de un proceso desde SWAP a la memoria
func dessuspenderProceso(pid int) error {
	// Verificar si existe la tabla de páginas
	tabla, existeTabla := tablasPaginas[pid]
	if !existeTabla {
		return fmt.Errorf("el proceso %d no tiene tabla de páginas", pid)
	}

	// Buscar todas las entradas de SWAP para este proceso
	swapMutex.Lock()
	paginasEnSwap := []int{}
	for _, entrada := range mapaSwap {
		if entrada.PID == pid && entrada.EnUso {
			paginasEnSwap = append(paginasEnSwap, entrada.Pagina)
		}
	}
	swapMutex.Unlock()

	// Si no hay páginas en SWAP, no hay nada que hacer
	if len(paginasEnSwap) == 0 {
		utils.InfoLog.Info("No hay páginas en SWAP para dessuspender", "pid", pid)
		return nil
	}

	// Verificar si hay suficientes marcos libres
	marcosNecesarios := len(paginasEnSwap)
	marcosDisponibles := contarMarcosLibres()
	if marcosDisponibles < marcosNecesarios {
		return fmt.Errorf("no hay suficientes marcos libres para dessuspender el proceso %d: "+
			"necesita %d, disponibles %d", pid, marcosNecesarios, marcosDisponibles)
	}

	// Asignar marcos para cada página
	marcosAsignados := []int{}
	for _, numPagina := range paginasEnSwap {
		// Asignar un nuevo marco
		marco, err := asignarMarco(pid)
		if err != nil {
			// Limpiar los marcos ya asignados y retornar error
			for _, m := range marcosAsignados {
				marcosLibres[m] = true
			}
			return fmt.Errorf("error asignando marco: %v", err)
		}
		marcosAsignados = append(marcosAsignados, marco)

		// Recuperar desde SWAP
		err = recuperarDeSwap(pid, numPagina, marco)
		if err != nil {
			// Limpiar los marcos ya asignados y retornar error
			for _, m := range marcosAsignados {
				marcosLibres[m] = true
			}
			return fmt.Errorf("error recuperando de SWAP: %v", err)
		}

		// Actualizar tabla de páginas
		actualizarTablaPaginas(pid, tabla, numPagina, marco, 1)
	}

	// Guardar la lista de marcos asignados al proceso
	marcosAsignadosPorProceso[pid] = marcosAsignados

	utils.InfoLog.Info(fmt.Sprintf("## PID: %d - Proceso dessuspendido desde SWAP", pid))

	return nil
}
