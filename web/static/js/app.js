// ─── Configurações Iniciais ─────────────────────────────────────────────────
const initialLat = -31.770687426923516;
const initialLng = -52.34135057529372;

// Limites de Pelotas-RS (impede arrastar o mapa pro mar)
const pelotasBounds = L.latLngBounds(
    L.latLng(-31.85, -52.45),
    L.latLng(-31.65, -52.20)
);

// ─── Inicialização do Mapa ────────────────────────────────────────────────────
const map = L.map('map', {
    zoomControl: false,
    maxBounds: pelotasBounds,
    maxBoundsViscosity: 1.0,
    minZoom: 12
}).setView([initialLat, initialLng], 15);

L.control.zoom({ position: 'topright' }).addTo(map);

L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/dark_all/{z}/{x}/{y}{r}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
    subdomains: 'abcd',
    maxZoom: 20
}).addTo(map);

// ─── Estado do App ────────────────────────────────────────────────────────────
const state = {
    drivers: {},
    selectedDriverId: null
};

// ─── Marcador do Restaurante ──────────────────────────────────────────────────
const restaurantIcon = L.divIcon({
    className: '',
    html: `<div style="
        width: 48px; height: 48px;
        background: #f97316;
        border-radius: 50% 50% 50% 0;
        transform: rotate(-45deg);
        border: 3px solid rgba(255,255,255,0.9);
        box-shadow: 0 4px 16px rgba(249,115,22,0.6);
        display: flex; align-items: center; justify-content: center;
    "><span style="transform: rotate(45deg); font-size: 20px; line-height:1;">🍽️</span></div>`,
    iconSize: [48, 48],
    iconAnchor: [24, 44],
    tooltipAnchor: [0, -42]
});

const restaurantMarker = L.marker([initialLat, initialLng], { icon: restaurantIcon, zIndexOffset: 1000 }).addTo(map);
restaurantMarker.bindTooltip('🏢 Restaurante Rei Julian', {
    permanent: false,
    direction: 'top',
    className: 'marker-label'
});

// ─── Despacho Dinâmico (clique direito no mapa) ───────────────────────────────
map.on('contextmenu', function(e) {
    const sel = document.getElementById('driver-select');
    const driverInput = document.getElementById('new-driver-name');
    if (!sel || !driverInput) return;

    if (sel.value === 'NEW') {
        alert('Por favor, crie o motoboy primeiro clicando no botão "Criar"!');
        return;
    }
    
    driverName = sel.value;

    if (!driverName) {
        alert('Por favor, selecione ou digite o nome do motoboy antes de clicar no mapa para despachar!');
        return;
    }

    fetch('/api/simulation/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            driver_name: driverName,
            dest_lat: e.latlng.lat,
            dest_lon: e.latlng.lng
        })
    }).then(async res => {
        if (!res.ok) {
            const errText = await res.text();
            throw new Error(errText || 'Falha no servidor');
        }
        return res.json();
    }).then(data => {
        driverInput.value = '';
    }).catch(err => {
        alert('Erro ao despachar: ' + err.message);
    });
});

// ─── Elementos do DOM ─────────────────────────────────────────────────────────
const listActive    = document.getElementById('driver-list-active');
const listDone      = document.getElementById('driver-list-done');
const titleActive   = document.getElementById('active-section-title');
const titleDone     = document.getElementById('done-section-title');

function updateSectionVisibility() {
    const hasActive = listActive && listActive.children.length > 0;
    const hasDone   = listDone   && listDone.children.length   > 0;
    if (titleActive) titleActive.style.display = hasActive ? 'flex' : 'none';
    if (titleDone)   titleDone.style.display   = hasDone   ? 'flex' : 'none';

    const emptyState = document.getElementById('empty-state');
    if (emptyState) emptyState.style.display = (hasActive || hasDone) ? 'none' : 'block';
}

// ─── Conexão SSE (posições em tempo real) ────────────────────────────────────
const source = new EventSource('/stream');

source.onmessage = function(event) {
    const data = JSON.parse(event.data);
    updateDriver(data);
};

source.onerror = function(err) {
    console.error('Erro na conexão SSE:', err);
};

// ─── Carregar drivers do banco ao iniciar ─────────────────────────────────────
function loadDrivers() {
    fetch('/api/drivers?t=' + Date.now())
        .then(res => res.json())
        .then(drivers => {
            if (!drivers) return;
            const sel = document.getElementById('driver-select');
            drivers.forEach(d => {
                if (sel && ![...sel.options].some(opt => opt.value === d.name)) {
                    const opt = document.createElement('option');
                    opt.value = d.name;
                    opt.textContent = d.name;
                    sel.insertBefore(opt, sel.lastElementChild);
                }

                // A API já filtra quem tem coordenadas válidas.
                // Mesmo assim, double-check para não duplicar.
                if (!state.drivers[d.name]) {
                    const pos = [d.last_lat, d.last_lon];
                    createDriver(d.name, pos, pos);
                    updateCardStatus(d.name, d.last_lat, d.last_lon, 0, d.status);
                }
            });
        })
        .catch(err => console.error('Erro ao buscar motoristas:', err));
}

loadDrivers();

// ─── Atualização via SSE ──────────────────────────────────────────────────────
function updateDriver(data) {
    const { id, token, latitude, longitude, target_lat, target_lon, step, status } = data;
    const pos = [latitude, longitude];
    const targetPos = [target_lat, target_lon];

    if (!state.drivers[id]) {
        createDriver(id, pos, targetPos);
    }

    const driver = state.drivers[id];
    if (token) driver.token = token;

    // Atualiza marcador
    driver.marker.setLatLng(pos);

    // Gerencia pin de destino de acordo com o status
    if (status === 'delivered') {
        // Entregou: remove o pin do destino
        if (map.hasLayer(driver.targetMarker)) map.removeLayer(driver.targetMarker);
        if (map.hasLayer(driver.targetPath))   map.removeLayer(driver.targetPath);
    } else if (status === 'returning' || status === 'available') {
        // Voltando: atualiza o pin para apontar para o restaurante
        driver.targetPath.setLatLngs([pos, targetPos]);
        if (!map.hasLayer(driver.targetMarker)) driver.targetMarker.addTo(map);
        driver.targetMarker.setLatLng(targetPos);
    } else {
        // in_route: pin de destino normal
        driver.targetPath.setLatLngs([pos, targetPos]);
        if (!map.hasLayer(driver.targetMarker)) driver.targetMarker.addTo(map);
        driver.targetMarker.setLatLng(targetPos);
    }

    // Atualiza rastro do histórico em memória
    driver.history.push(pos);
    if (driver.history.length > 200) driver.history.shift();
    driver.path.setLatLngs(driver.history);

    // Atualiza card da sidebar
    updateCardStatus(id, latitude, longitude, step, status);

    // Segue o selecionado
    if (state.selectedDriverId === id) {
        map.panTo(pos);
    }
}

// ─── Criar driver no mapa e na sidebar ───────────────────────────────────────
function createDriver(id, pos, targetPos) {
    const sel = document.getElementById('driver-select');
    if (sel && ![...sel.options].some(opt => opt.value === id)) {
        const opt = document.createElement('option');
        opt.value = id;
        opt.textContent = id;
        sel.insertBefore(opt, sel.lastElementChild);
    }

    const driverColor = getDriverColor(id);

    // 1. Marcador do entregador
    const marker = L.marker(pos).addTo(map);
    marker.bindTooltip(id, {
        permanent: true,
        direction: 'top',
        className: 'marker-label'
    }).openTooltip();
    marker.on('click', () => selectDriver(id));

    // 2. Pin de destino — visível desde o início
    const deliveryPinIcon = L.divIcon({
        className: '',
        html: `
            <div style="
                position: relative;
                width: 36px;
                height: 36px;
            ">
                <div style="
                    position: absolute;
                    inset: 0;
                    border-radius: 50%;
                    background: ${driverColor};
                    opacity: 0.18;
                    animation: pulse-ring 1.8s ease-out infinite;
                "></div>
                <div style="
                    position: absolute;
                    top: 50%; left: 50%;
                    transform: translate(-50%, -50%);
                    width: 20px; height: 20px;
                    background: ${driverColor};
                    border-radius: 50% 50% 50% 0;
                    transform: translate(-50%, -60%) rotate(-45deg);
                    box-shadow: 0 3px 10px rgba(0,0,0,0.5);
                    border: 2px solid rgba(255,255,255,0.85);
                "></div>
                <div style="
                    position: absolute;
                    top: 50%; left: 50%;
                    width: 7px; height: 7px;
                    background: white;
                    border-radius: 50%;
                    transform: translate(-50%, -85%);
                "></div>
            </div>`,
        iconSize: [36, 36],
        iconAnchor: [18, 32],
        tooltipAnchor: [0, -28]
    });
    const targetMarkerLayer = L.marker(targetPos, { icon: deliveryPinIcon }).addTo(map);
    targetMarkerLayer.bindTooltip(`📦 Destino de ${id}`, {
        permanent: false,
        direction: 'top',
        className: 'marker-label'
    });

    // 3. Rastro do histórico — ESCONDIDO até o driver ser selecionado
    const path = L.polyline([], {
        color: driverColor,
        weight: 4,
        opacity: 0.85,
        lineCap: 'round',
        lineJoin: 'round'
    });

    // 4. Linha reta até o alvo — ESCONDIDA por padrão
    const targetPathLayer = L.polyline([pos, targetPos], {
        color: driverColor,
        weight: 3,
        opacity: 0.55,
        dashArray: '8, 10',
        lineCap: 'round'
    });

    // 5. Card na sidebar → vai para a seção "Frota Ativa"
    const card = document.createElement('div');
    card.className = 'driver-card';
    card.id = `card-${id}`;
    card.innerHTML = `
        <div class="driver-info">
            <span class="driver-name">${id}</span>
            <div class="client-actions" style="display:none; align-items:center; gap:5px;">
                <button class="btn-copy-link" style="background:transparent; border:none; cursor:pointer; font-size:1.1rem;" title="Copiar Link" onclick="event.stopPropagation(); copyClientLink('${id}')">📋</button>
                <a href="#" target="_blank" class="client-link" style="font-size:1.1rem; text-decoration:none;" title="Abrir Link" onclick="event.stopPropagation()">🔗</a>
            </div>
            <span class="status-badge status-online">Em Rota</span>
        </div>
        <div class="driver-details">
            <span class="coordinate-label">LAT: <span class="lat-val">—</span></span>
            <span class="coordinate-label">LON: <span class="lon-val">—</span></span>
            <div style="margin-top:10px;font-size:0.75rem;color:#999; display: flex; justify-content: space-between; align-items: center;">
                <span>Atualizações: <span class="step-val">0</span></span>
            </div>
        </div>
    `;
    card.onclick = () => selectDriver(id);
    if (listActive) listActive.appendChild(card);
    updateSectionVisibility();

    state.drivers[id] = {
        marker,
        path,
        targetMarker: targetMarkerLayer,
        targetPath: targetPathLayer,
        card,
        history: [pos],
        color: driverColor,
        showingTarget: false,
        token: null
    };

    // Carregar histórico completo do banco
    fetch(`/api/history/${id}`)
        .then(res => res.json())
        .then(data => {
            if (data && data.length > 0) {
                const historyPos = data.map(p => [p.latitude, p.longitude]);
                state.drivers[id].history = historyPos;
                path.setLatLngs(historyPos);
            }
        })
        .catch(err => console.error('Erro ao carregar histórico:', err));
}

// ─── Atualiza card da sidebar com dados atuais ────────────────────────────────
function updateCardStatus(id, lat, lon, step, status) {
    const driver = state.drivers[id];
    if (!driver) return;
    const card = driver.card;

    const latStr = (typeof lat === 'number' && isFinite(lat)) ? lat.toFixed(6) : '—';
    const lonStr = (typeof lon === 'number' && isFinite(lon)) ? lon.toFixed(6) : '—';

    card.querySelector('.lat-val').textContent = latStr;
    card.querySelector('.lon-val').textContent = lonStr;
    card.querySelector('.step-val').textContent = step ?? 0;

    const badge = card.querySelector('.status-badge');
    
    const clientActions = card.querySelector('.client-actions');
    const clientLink = card.querySelector('.client-link');
    
    if (driver.token && status === 'in_route') {
        clientActions.style.display = 'flex';
        clientLink.href = '/client/?token=' + driver.token;
    } else {
        clientActions.style.display = 'none';
    }

    if (status === 'delivered') {
        badge.textContent = '✅ Entregue';
        badge.className = 'status-badge';
        badge.style.background = 'rgba(244, 63, 94, 0.15)';
        badge.style.color = '#f43f5e';
        badge.style.border = '1px solid rgba(244, 63, 94, 0.3)';
        // O card continua na lista ativa se houver fila de entregas, mas por hora vamos movê-lo visualmente
        // sem deselecionar o motoboy do mapa.
        if (listDone && card.parentElement !== listDone) {
            listDone.appendChild(card);
            updateSectionVisibility();
        }
    } else if (status === 'returning') {
        badge.textContent = '↩️ Retornando';
        badge.className = 'status-badge';
        badge.style.background = 'rgba(251, 146, 60, 0.15)';
        badge.style.color = '#fb923c';
        badge.style.border = '1px solid rgba(251, 146, 60, 0.3)';
        // Volta para frota ativa se estava nas concluídas
        if (listActive && card.parentElement !== listActive) {
            listActive.appendChild(card);
            updateSectionVisibility();
        }
    } else if (status === 'available') {
        badge.textContent = '🏠 Na Base';
        badge.className = 'status-badge';
        badge.style.background = 'rgba(129, 140, 248, 0.15)';
        badge.style.color = '#818cf8';
        badge.style.border = '1px solid rgba(129, 140, 248, 0.3)';
        // Mantém na frota ativa
        if (listActive && card.parentElement !== listActive) {
            listActive.appendChild(card);
            updateSectionVisibility();
        }
    } else if (status === 'offline') {
        badge.textContent = 'Offline';
        badge.className = 'status-badge status-offline';
        badge.style = '';
    } else {
        // in_route
        badge.textContent = '🛵 Em Rota';
        badge.className = 'status-badge status-online';
        badge.style = '';
        // Garante que está na lista ativa
        if (listActive && card.parentElement !== listActive) {
            listActive.appendChild(card);
            updateSectionVisibility();
        }
    }
}

// ─── Selecionar driver (destaca card e exibe rastro) ─────────────────────────
function selectDriver(id) {
    // Remove destaque e rastro do driver anterior
    if (state.selectedDriverId && state.drivers[state.selectedDriverId]) {
        const prev = state.drivers[state.selectedDriverId];
        prev.card.classList.remove('active');
        if (map.hasLayer(prev.path)) map.removeLayer(prev.path);
    }

    state.selectedDriverId = id;
    const driver = state.drivers[id];
    driver.card.classList.add('active');

    // Exibe rastro apenas do driver selecionado
    if (!map.hasLayer(driver.path)) driver.path.addTo(map);

    map.flyTo(driver.marker.getLatLng(), 16);
}

// ─── Toggle destino (linha tracejada + marcador alvo) ────────────────────────
function toggleTarget(id) {
    const driver = state.drivers[id];
    driver.showingTarget = !driver.showingTarget;

    if (driver.showingTarget) {
        driver.targetPath.addTo(map);
        driver.targetMarker.addTo(map);
        const bounds = L.latLngBounds([driver.marker.getLatLng(), driver.targetMarker.getLatLng()]);
        map.fitBounds(bounds, { padding: [50, 50] });
    } else {
        map.removeLayer(driver.targetPath);
        map.removeLayer(driver.targetMarker);
    }
}

// ─── Cor determinística por ID (evita cores trocando no reload) ────────────
function getDriverColor(id) {
    const colors = ['#818cf8', '#34d399', '#fb923c', '#f472b6', '#60a5fa', '#a78bfa', '#facc15'];
    let hash = 0;
    for (let i = 0; i < id.length; i++) {
        hash = id.charCodeAt(i) + ((hash << 5) - hash);
    }
    return colors[Math.abs(hash) % colors.length];
}

window.copyClientLink = function(id) {
    const driver = state.drivers[id];
    if (driver && driver.token) {
        const url = window.location.origin + '/client/?token=' + driver.token;
        navigator.clipboard.writeText(url).then(() => {
            // Mostrar pequeno feedback visual
            const btn = document.querySelector(`#card-${id} .btn-copy-link`);
            if (btn) {
                const oldText = btn.textContent;
                btn.textContent = '✅';
                setTimeout(() => btn.textContent = oldText, 1500);
            }
        }).catch(err => {
            console.error('Falha ao copiar:', err);
        });
    }
}

