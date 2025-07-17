#!/bin/bash

echo "=== Deteniendo Prueba de Estabilidad General ==="

# Leer PIDs guardados
if [ -f .test_pids ]; then
    source .test_pids
    
    echo "🛑 Deteniendo procesos..."
    
    # Detener en orden inverso
    [ ! -z "$IO4_PID" ] && kill $IO4_PID 2>/dev/null && echo "  ✓ IO4 detenido"
    [ ! -z "$IO3_PID" ] && kill $IO3_PID 2>/dev/null && echo "  ✓ IO3 detenido" 
    [ ! -z "$IO2_PID" ] && kill $IO2_PID 2>/dev/null && echo "  ✓ IO2 detenido"
    [ ! -z "$IO1_PID" ] && kill $IO1_PID 2>/dev/null && echo "  ✓ IO1 detenido"
    
    [ ! -z "$CPU4_PID" ] && kill $CPU4_PID 2>/dev/null && echo "  ✓ CPU4 detenido"
    [ ! -z "$CPU3_PID" ] && kill $CPU3_PID 2>/dev/null && echo "  ✓ CPU3 detenido"
    [ ! -z "$CPU2_PID" ] && kill $CPU2_PID 2>/dev/null && echo "  ✓ CPU2 detenido"
    [ ! -z "$CPU1_PID" ] && kill $CPU1_PID 2>/dev/null && echo "  ✓ CPU1 detenido"
    
    [ ! -z "$KERNEL_PID" ] && kill $KERNEL_PID 2>/dev/null && echo "  ✓ Kernel detenido"
    [ ! -z "$MEMORIA_PID" ] && kill $MEMORIA_PID 2>/dev/null && echo "  ✓ Memoria detenido"
    
    # Limpiar archivo de PIDs
    rm -f .test_pids
else
    echo "⚠️  No se encontró archivo .test_pids"
    echo "🔍 Buscando procesos manualmente..."
    
    # Buscar y matar procesos por nombre
    pkill -f "./bin/memoria" && echo "  ✓ Memoria detenido"
    pkill -f "./bin/kernel" && echo "  ✓ Kernel detenido"  
    pkill -f "./bin/cpu" && echo "  ✓ CPUs detenidos"
    pkill -f "./bin/io" && echo "  ✓ IOs detenidos"
fi

echo ""
echo "✅ Prueba detenida"
echo ""
echo "📊 Para analizar resultados:"
echo "   ./analyze_test.sh"