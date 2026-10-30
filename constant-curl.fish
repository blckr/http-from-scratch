#!/usr/bin/env fish

set -l base_url "http://localhost:8080"
set -l endpoints /test /delay /flaky /payload /unknown

echo "=== Verifiziere Verbindung zu $base_url ==="
if not curl -s --connect-timeout 1 "$base_url/test" >/dev/null
    echo "Fehler: $base_url ist nicht erreichbar!"
    echo "Bitte stelle sicher, dass 'make forward-app' im Hintergrund laeuft."
    exit 1
end

echo "=== Raffinierter Traffic-Generator gestartet (Ctrl+C zum Stoppen) ==="

# Hilfsfunktion für zufällige Hex-Strings (W3C Tracing)
function gen_hex
    head -c $argv[1] /dev/urandom | xxd -p | tr -d '\n'
end

# Hilfsfunktion zum Absetzen eines getrackten Requests
function fire_request
    set -l base_url $argv[1]
    set -l endpoint $argv[2]
    set -l trace_id $argv[3]
    set -l span_id $argv[4]

    set -l traceparent "00-$trace_id-$span_id-01"

    curl -s -o /dev/null -w "%{http_code}" \
        -H "traceparent: $traceparent" \
        -H "x-client-simulation: loadgen" \
        "$base_url$endpoint" &
end

set -l tick 0
set -l total_sent 0

while true
    set tick (math $tick + 1)

    # 1. Sinusförmige Grundlast berechnen (Schwankt harmonisch zwischen 2 und 8 Requests)
    # math in fish unterstuetzt sin()
    set -l sine_wave (math "round((sin($tick / 10) + 1) * 3) + 2")
    set -l batch_size $sine_wave

    # 2. Event-Würfel rollen (1 bis 100)
    set -l roll (random 1 100)
    set -l event_desc "Normal (Welle: $batch_size reqs)"

    # --- EVENT 1: DDoS / Traffic Spike (4% Chance) ---
    if test $roll -ge 97
        set batch_size 30
        set event_desc "⚡ SPIKE: 30 parallele Requests auf /test & /delay"
        set -l trace_id (gen_hex 16)

        for j in (seq 1 $batch_size)
            set -l endpoint (random choice "/test" "/delay")
            set -l span_id (gen_hex 8)
            fire_request $base_url $endpoint $trace_id $span_id
        end
        wait

        # --- EVENT 2: Chaos Storm / Error Wave (5% Chance) ---
    else if test $roll -ge 92
        set batch_size 16
        set event_desc "🔥 CHAOS: Error-Welle auf /flaky (500) und /unknown (404)"
        set -l trace_id (gen_hex 16)

        for j in (seq 1 $batch_size)
            set -l endpoint (random choice "/flaky" "/unknown")
            set -l span_id (gen_hex 8)
            fire_request $base_url $endpoint $trace_id $span_id
        end
        wait

        # --- EVENT 3: Payload Flood (4% Chance) ---
    else if test $roll -ge 88
        set batch_size 10
        set event_desc "📦 BANDWIDTH: Heavy Payload Burst auf /payload"
        set -l trace_id (gen_hex 16)

        for j in (seq 1 $batch_size)
            set -l span_id (gen_hex 8)
            fire_request $base_url /payload $trace_id $span_id
        end
        wait

        # --- EVENT 4: Client Network Dropout (2% Chance) ---
    else if test $roll -le 2
        set event_desc "🛑 DROPOUT: Simulierter Netzwerk-Ausfall (Pause 5s)"
        echo "[Tick $tick] $event_desc"
        sleep 5
        continue

        # --- NORMALER TRAFFIC (Sinus-Welle) ---
    else
        set -l trace_id (gen_hex 16)
        for j in (seq 1 $batch_size)
            set -l endpoint (random choice $endpoints)
            set -l span_id (gen_hex 8)
            fire_request $base_url $endpoint $trace_id $span_id
        end
        wait
    end

    set total_sent (math $total_sent + $batch_size)
    echo "[Tick $tick | Gesamt: $total_sent] $event_desc"

    # Kleiner dynamischer Jitter im Delay (120ms bis 220ms)
    set -l jitter (math (random 120 220) / 1000)
    sleep $jitter
end
