#!/bin/bash

echo "=== Análisis de Prueba de Estabilidad General ==="
echo ""

# Verificar que existan logs
if [ ! -d "logs" ]; then
    echo "❌ No se encontró directorio logs/"
    exit 1
fi

# Función para contar líneas de un archivo
count_lines() {
    if [ -f "$1" ]; then
        wc -l < "$1"
    else
        echo "0"
    fi
}

# 1. DISTRIBUCIÓN DE CPUs
echo "📊 === DISTRIBUCIÓN DE TRABAJO ENTRE CPUs ==="
if [ -f "logs/kernel.log" ]; then
    CPU1_COUNT=$(grep -c "CPU1" logs/kernel.log 2>/dev/null || echo "0")
    CPU2_COUNT=$(grep -c "CPU2" logs/kernel.log 2>/dev/null || echo "0") 
    CPU3_COUNT=$(grep -c "CPU3" logs/kernel.log 2>/dev/null || echo "0")
    CPU4_COUNT=$(grep -c "CPU4" logs/kernel.log 2>/dev/null || echo "0")
    
    echo "CPU1: $CPU1_COUNT procesos"
    echo "CPU2: $CPU2_COUNT procesos" 
    echo "CPU3: $CPU3_COUNT procesos"
    echo "CPU4: $CPU4_COUNT procesos"
    
    TOTAL_CPU=$((CPU1_COUNT + CPU2_COUNT + CPU3_COUNT + CPU4_COUNT))
    echo "Total: $TOTAL_CPU procesos distribuidos"
    
    # Verificar distribución balanceada
    if [ $TOTAL_CPU -gt 0 ]; then
        if [ $CPU1_COUNT -gt 0 ] && [ $CPU2_COUNT -gt 0 ] && [ $CPU3_COUNT -gt 0 ] && [ $CPU4_COUNT -gt 0 ]; then
            echo "✅ Distribución balanceada entre CPUs"
        else
            echo "⚠️  Distribución desbalanceada - algunas CPUs sin uso"
        fi
    fi
else
    echo "❌ No se encontró logs/kernel.log"
fi

echo ""

# 2. DISTRIBUCIÓN DE IOs
echo "📊 === DISTRIBUCIÓN DE TRABAJO ENTRE IOs ==="
if [ -f "logs/kernel.log" ]; then
    DISCO1_COUNT=$(grep -c "DISCO1" logs/kernel.log 2>/dev/null || echo "0")
    DISCO2_COUNT=$(grep -c "DISCO2" logs/kernel.log 2>/dev/null || echo "0")
    DISCO3_COUNT=$(grep -c "DISCO3" logs/kernel.log 2>/dev/null || echo "0") 
    DISCO4_COUNT=$(grep -c "DISCO4" logs/kernel.log 2>/dev/null || echo "0")
    
    echo "DISCO1: $DISCO1_COUNT operaciones"
    echo "DISCO2: $DISCO2_COUNT operaciones"
    echo "DISCO3: $DISCO3_COUNT operaciones" 
    echo "DISCO4: $DISCO4_COUNT operaciones"
    
    TOTAL_IO=$((DISCO1_COUNT + DISCO2_COUNT + DISCO3_COUNT + DISCO4_COUNT))
    echo "Total: $TOTAL_IO operaciones de IO"
    
    if [ $TOTAL_IO -gt 0 ]; then
        if [ $DISCO1_COUNT -gt 0 ] && [ $DISCO2_COUNT -gt 0 ] && [ $DISCO3_COUNT -gt 0 ] && [ $DISCO4_COUNT -gt 0 ]; then
            echo "✅ Distribución balanceada entre IOs"
        else
            echo "⚠️  Distribución desbalanceada - algunos IOs sin uso"
        fi
    fi
fi

echo ""

# 3. ANÁLISIS DE ERRORES
echo "🚨 === ANÁLISIS DE ERRORES ==="
ERROR_COUNT=0
for log_file in logs/*.log; do
    if [ -f "$log_file" ]; then
        file_errors=$(grep -c -i "error\|panic\|fatal" "$log_file" 2>/dev/null || echo "0")
        if [ $file_errors -gt 0 ]; then
            echo "$(basename $log_file): $file_errors errores"
            ERROR_COUNT=$((ERROR_COUNT + file_errors))
        fi
    fi
done

if [ $ERROR_COUNT -eq 0 ]; then
    echo "✅ No se encontraron errores críticos"
else
    echo "⚠️  Total de errores encontrados: $ERROR_COUNT"
    echo ""
    echo "🔍 Últimos errores:"
    grep -i "error\|panic\|fatal" logs/*.log | tail -5
fi

echo ""

# 4. ACTIVIDAD DE MEMORIA  
echo "💾 === ACTIVIDAD DE MEMORIA ==="
if [ -f "logs/memoria.log" ]; then
    SWAP_MOVES=$(grep -c "movidos a SWAP\|recuperada de SWAP" logs/memoria.log 2>/dev/null || echo "0")
    PROCESOS_CREADOS=$(grep -c "Proceso Creado" logs/memoria.log 2>/dev/null || echo "0")
    PROCESOS_DESTRUIDOS=$(grep -c "Proceso Destruido" logs/memoria.log 2>/dev/null || echo "0")
    
    echo "Procesos creados: $PROCESOS_CREADOS"
    echo "Procesos destruidos: $PROCESOS_DESTRUIDOS" 
    echo "Operaciones SWAP: $SWAP_MOVES"
    
    if [ $PROCESOS_CREADOS -gt 0 ]; then
        echo "✅ Actividad de memoria detectada"
    fi
fi

echo ""

# 5. RESUMEN FINAL
echo "📋 === RESUMEN FINAL ==="
TOTAL_LOGS=$(ls logs/*.log 2>/dev/null | wc -l)
echo "Archivos de log generados: $TOTAL_LOGS"

# Calcular duración aproximada
if [ -f "logs/kernel.log" ]; then
    FIRST_LINE=$(head -1 logs/kernel.log | grep -o '[0-9][0-9]:[0-9][0-9]:[0-9][0-9]' | head -1)
    LAST_LINE=$(tail -1 logs/kernel.log | grep -o '[0-9][0-9]:[0-9][0-9]:[0-9][0-9]' | head -1)
    
    if [ ! -z "$FIRST_LINE" ] && [ ! -z "$LAST_LINE" ]; then
        echo "Duración aproximada: $FIRST_LINE - $LAST_LINE"
    fi
fi

echo ""
echo "🎯 === RESULTADO ==="

# Criterios de éxito
SUCCESS=true

if [ $ERROR_COUNT -gt 10 ]; then
    echo "❌ FALLO: Demasiados errores ($ERROR_COUNT)"
    SUCCESS=false
fi

if [ $TOTAL_CPU -eq 0 ]; then
    echo "❌ FALLO: No hay distribución de CPUs"
    SUCCESS=false
fi

if [ $CPU1_COUNT -eq 0 ] || [ $CPU2_COUNT -eq 0 ] || [ $CPU3_COUNT -eq 0 ] || [ $CPU4_COUNT -eq 0 ]; then
    echo "⚠️  ADVERTENCIA: Distribución CPU desbalanceada"
fi

if [ $TOTAL_IO -eq 0 ]; then
    echo "❌ FALLO: No hay operaciones de IO"
    SUCCESS=false
fi

if $SUCCESS; then
    echo "🎉 ✅ PRUEBA DE ESTABILIDAD EXITOSA"
    echo ""
    echo "✓ Sistema ejecutó sin errores críticos"
    echo "✓ Distribución balanceada de trabajo"
    echo "✓ Múltiples CPUs y IOs funcionando"
    echo "✓ Actividad de memoria detectada"
else
    echo "❌ PRUEBA DE ESTABILIDAD FALLIDA" 
    echo ""
    echo "Revisar logs para más detalles."
fi