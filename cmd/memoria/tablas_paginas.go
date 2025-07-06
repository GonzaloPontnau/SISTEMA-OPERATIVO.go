package main

import (
	"fmt"
	"sync"


	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

// Variables globales
var tablasMemoria = make(map[int]*TablaPaginas)
var ultimoID = 0
var mutexTablas = &sync.Mutex{}

// Función para crear tabla de páginas para un nuevo proceso
func crearTablasPaginas(pid int, tamanio int) (*TablaPaginas, error) {
	// Calcular número de páginas necesarias
	numPaginas := calcularNumeroPaginas(tamanio)

	// Verificar si hay suficientes marcos libres
	marcosFree := contarMarcosLibres()
	if marcosFree < numPaginas {
		return nil, fmt.Errorf("no hay suficientes marcos libres (%d) para el proceso %d que requiere %d páginas",
			marcosFree, pid, numPaginas)
	}

	// Crear tabla de nivel 1
	tablaNivel1 := &TablaPaginas{
		Entradas: make([]EntradaTabla, config.EntriesPerPage),
		Nivel:    1,
	}

	// Inicializar las entradas como inválidas
	for i := range tablaNivel1.Entradas {
		tablaNivel1.Entradas[i].Valido = false
	}

	// Asignar tabla al proceso
	tablasPaginas[pid] = tablaNivel1

	// Registrar que este proceso necesita estas páginas
	marcosAsignadosPorProceso[pid] = []int{}

	utils.InfoLog.Info("Tabla de páginas creada", "PID", pid, "páginas", numPaginas)

	return tablaNivel1, nil
}

// Navega recursivamente los niveles de tablas para obtener el marco final
func obtenerMarcoDesdeTabla(pid int, tabla *TablaPaginas, numPagina int, nivelActual int) (int, error) {
	// Actualizar métricas de acceso a tablas de páginas
	actualizarMetricasAccesoTabla(pid)

	// Calcular índice en el nivel actual
	indice := calcularIndiceEnNivel(numPagina, nivelActual)

	// Verificar si la entrada es válida
	if indice >= len(tabla.Entradas) || !tabla.Entradas[indice].Valido {
		// Si no es válida, necesitamos crear estructuras adicionales
		if nivelActual < config.NumberOfLevels {
			// Crear tabla para el siguiente nivel
			nuevaTabla := crearTablaSiguienteNivel(pid, tabla, indice, nivelActual+1)
			return obtenerMarcoDesdeTabla(pid, nuevaTabla, numPagina, nivelActual+1)
		} else {
			// En el último nivel, necesitamos asignar un marco
			marco, err := asignarMarco(pid)
			if err != nil {
				return 0, err
			}

			// Actualizar la entrada
			tabla.Entradas[indice].Marco = marco
			tabla.Entradas[indice].Presente = true
			tabla.Entradas[indice].Valido = true

			return marco, nil
		}
	}

	// Si estamos en el último nivel, devolver el marco
	if nivelActual == config.NumberOfLevels {
		if !tabla.Entradas[indice].Presente {
			// Traer página de SWAP si es necesario
			err := traerPaginaDeSwap(pid, numPagina, tabla.Entradas[indice].Marco)
			if err != nil {
				return 0, err
			}
			tabla.Entradas[indice].Presente = true
		}
		return tabla.Entradas[indice].Marco, nil
	}

	// Si no estamos en el último nivel, obtener la siguiente tabla
	siguienteTabla := obtenerTablaSiguienteNivel(tabla.Entradas[indice].Direccion)

	// Llamada recursiva al siguiente nivel
	return obtenerMarcoDesdeTabla(pid, siguienteTabla, numPagina, nivelActual+1)
}

// Calcula el índice en un nivel específico de la tabla de páginas
func calcularIndiceEnNivel(numPagina int, nivel int) int {
	potencia := 1
	for i := 0; i < config.NumberOfLevels-nivel; i++ {
		potencia *= config.EntriesPerPage
	}
	return (numPagina / potencia) % config.EntriesPerPage
}

// Función auxiliar para crear una nueva tabla del siguiente nivel
func crearTablaSiguienteNivel(pid int, tablaActual *TablaPaginas, indice int, nuevoNivel int) *TablaPaginas {
	nuevaTabla := &TablaPaginas{
		Entradas: make([]EntradaTabla, config.EntriesPerPage),
		Nivel:    nuevoNivel,
	}

	// Inicializar entradas como inválidas
	for i := range nuevaTabla.Entradas {
		nuevaTabla.Entradas[i].Valido = false
	}

	// Asignar dirección en la tabla del nivel anterior
	tablaActual.Entradas[indice].Direccion = almacenarTablaEnMemoria(nuevaTabla)
	tablaActual.Entradas[indice].Valido = true
	tablaActual.Entradas[indice].Presente = true

	return nuevaTabla
}

// Almacena una tabla de páginas y devuelve un ID único
func almacenarTablaEnMemoria(tabla *TablaPaginas) int {
	mutexTablas.Lock()
	defer mutexTablas.Unlock()

	ultimoID++
	tablasMemoria[ultimoID] = tabla
	return ultimoID
}

// Obtiene una tabla usando su ID
func obtenerTablaSiguienteNivel(id int) *TablaPaginas {
	mutexTablas.Lock()
	defer mutexTablas.Unlock()

	return tablasMemoria[id]
}

// marcarPaginaNoPresente marca una página como no presente en la tabla de páginas
func marcarPaginaNoPresente(pid int, tabla *TablaPaginas, numPagina int, nivelActual int) {
	if nivelActual == config.NumberOfLevels {
		indice := calcularIndiceEnNivel(numPagina, nivelActual)
		if indice < len(tabla.Entradas) && tabla.Entradas[indice].Valido {
			tabla.Entradas[indice].Presente = false
		}
	} else {
		indice := calcularIndiceEnNivel(numPagina, nivelActual)
		if indice < len(tabla.Entradas) && tabla.Entradas[indice].Valido {
			siguienteTabla := obtenerTablaSiguienteNivel(tabla.Entradas[indice].Direccion)
			marcarPaginaNoPresente(pid, siguienteTabla, numPagina, nivelActual+1)
		}
	}
}

// actualizarTablaPaginas actualiza la entrada de la tabla para una página
func actualizarTablaPaginas(pid int, tabla *TablaPaginas, numPagina int, marco int, nivelActual int) {
	if nivelActual == config.NumberOfLevels {
		indice := calcularIndiceEnNivel(numPagina, nivelActual)
		if indice < len(tabla.Entradas) {
			tabla.Entradas[indice].Marco = marco
			tabla.Entradas[indice].Presente = true
			tabla.Entradas[indice].Valido = true
		}
	} else {
		indice := calcularIndiceEnNivel(numPagina, nivelActual)
		if indice < len(tabla.Entradas) {
			if !tabla.Entradas[indice].Valido {
				// Crear nueva tabla si no existe
				nuevaTabla := crearTablaSiguienteNivel(pid, tabla, indice, nivelActual+1)
				actualizarTablaPaginas(pid, nuevaTabla, numPagina, marco, nivelActual+1)
			} else {
				siguienteTabla := obtenerTablaSiguienteNivel(tabla.Entradas[indice].Direccion)
				actualizarTablaPaginas(pid, siguienteTabla, numPagina, marco, nivelActual+1)
			}
		}
	}
}

// encontrarPaginaPorMarco busca recursivamente qué página corresponde a un marco
func encontrarPaginaPorMarco(pid int, tabla *TablaPaginas, marco int, nivelActual int) int {
	if nivelActual == config.NumberOfLevels {
		// En el último nivel, buscamos el marco en las entradas
		for i, entrada := range tabla.Entradas {
			if entrada.Valido && entrada.Marco == marco {
				// Calculamos el número de página a partir del índice en el último nivel
				// Para simplificar, usamos una aproximación básica
				potencia := 1
				for j := 0; j < config.NumberOfLevels-nivelActual; j++ {
					potencia *= config.EntriesPerPage
				}
				return i * potencia
			}
		}
		return -1
	} else {
		// En niveles intermedios, buscamos en las tablas siguientes
		for _, entrada := range tabla.Entradas {
			if entrada.Valido {
				siguienteTabla := obtenerTablaSiguienteNivel(entrada.Direccion)
				numPagina := encontrarPaginaPorMarco(pid, siguienteTabla, marco, nivelActual+1)
				if numPagina != -1 {
					return numPagina
				}
			}
		}
		return -1
	}
}
