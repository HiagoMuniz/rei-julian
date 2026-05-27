// Configurações Iniciais
const initialLat = -31.770687426923516;
const initialLng = -52.34135057529372;

// Inicialização do Mapa
const map = L.map('map', {
    zoomControl: false // Vamos reposicionar o zoom para a direita
}).setView([initialLat, initialLng], 15);

L.control.zoom({ position: 'topright' }).addTo(map);

// Camada de Mapa (Estilo mais limpo)
L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
    subdomains: 'abcd',
    maxZoom: 20
}).addTo(map);

// Estado do App
const state = {
    drivers: {}, // { id: { marker, path, targetMarker, targetPath, card, history: [], color: str } }
    selectedDriverId: null,
    showPaths: true
};

// Elementos do DOM
const driverList = document.getElementById('driver-list');

// Conexão com o Servidor (SSE)
const source = new EventSource("/stream");

source.onmessage = function(event) {
    const data = JSON.parse(event.data);
    updateDriver(data);
};

source.onerror = function(err) {
    console.error("Erro na conexão SSE:", err);
};

function updateDriver(data) {
    const { id, latitude, longitude, target_lat, target_lon, step, timestamp } = data;
    const pos = [latitude, longitude];
    const targetPos = [target_lat, target_lon];

    if (!state.drivers[id]) {
        createDriver(id, pos, targetPos);
    }

    const driver = state.drivers[id];
    
    // Atualiza Marcador do Entregador
    driver.marker.setLatLng(pos);
    
    // Atualiza linha para o destino
    driver.targetPath.setLatLngs([pos, targetPos]);
    
    // Atualiza Histórico/Rastro
    driver.history.push(pos);
    if (driver.history.length > 50) driver.history.shift(); // Mantém os últimos 50 pontos
    driver.path.setLatLngs(driver.history);

    // Atualiza Card na Sidebar
    updateCard(id, latitude, longitude, step);

    // Se for o selecionado, segue no mapa
    if (state.selectedDriverId === id) {
        map.panTo(pos);
    }
}

function createDriver(id, pos, targetPos) {
    const driverColor = getRandomColor();

    // 1. Criar Marcador do Entregador
    const marker = L.marker(pos).addTo(map);
    marker.bindTooltip(id, {
        permanent: true,
        direction: 'top',
        className: 'marker-label'
    }).openTooltip();
    marker.on('click', () => selectDriver(id));

    // 2. Criar Marcador do Destino (Alvo)
    const targetMarkerLayer = L.circleMarker(targetPos, {
        color: driverColor,
        fillColor: driverColor,
        fillOpacity: 0.5,
        radius: 8
    }).addTo(map);
    targetMarkerLayer.bindTooltip(`Destino: ${id}`, {direction: 'bottom'});

    // 3. Criar Polylinha (Rastro do Histórico)
    const path = L.polyline([], {
        color: driverColor,
        weight: 3,
        opacity: 0.4,
        dashArray: '5, 10'
    }).addTo(map);

    // 4. Criar Polylinha (Linha reta até o alvo)
    const targetPathLayer = L.polyline([pos, targetPos], {
        color: driverColor,
        weight: 2,
        opacity: 0.8,
        dashArray: '10, 10'
    }).addTo(map);

    // Esconde por padrão para não poluir
    map.removeLayer(targetPathLayer);
    map.removeLayer(targetMarkerLayer);

    // 5. Criar Card na Sidebar
    const card = document.createElement('div');
    card.className = 'driver-card';
    card.id = `card-${id}`;
    card.innerHTML = `
        <div class="driver-info">
            <span class="driver-name">${id}</span>
            <span class="status-badge status-online">Em Rota</span>
        </div>
        <div class="driver-details">
            <div style="display: flex; gap: 10px; margin-bottom: 8px;">
                <button class="btn-small" onclick="event.stopPropagation(); toggleTarget('${id}')">Ver Rota</button>
            </div>
            <span class="coordinate-label">LAT: <span class="lat-val">0</span></span>
            <span class="coordinate-label">LON: <span class="lon-val">0</span></span>
            <div style="margin-top: 10px; font-size: 0.75rem; color: #999;">
                Atualizações: <span class="step-val">0</span>
            </div>
        </div>
    `;
    card.onclick = () => selectDriver(id);
    driverList.appendChild(card);

    state.drivers[id] = {
        marker: marker,
        path: path,
        targetMarker: targetMarkerLayer,
        targetPath: targetPathLayer,
        card: card,
        history: [pos],
        color: driverColor,
        showingTarget: false
    };
}

function updateCard(id, lat, lon, step) {
    const card = state.drivers[id].card;
    card.querySelector('.lat-val').textContent = lat.toFixed(6);
    card.querySelector('.lon-val').textContent = lon.toFixed(6);
    card.querySelector('.step-val').textContent = step;
}

function selectDriver(id) {
    // Desmarcar anterior
    if (state.selectedDriverId) {
        const prevCard = state.drivers[state.selectedDriverId].card;
        prevCard.classList.remove('active');
    }

    // Marcar novo
    state.selectedDriverId = id;
    const driver = state.drivers[id];
    driver.card.classList.add('active');
    
    // Centralizar
    map.flyTo(driver.marker.getLatLng(), 16);
}

function toggleTarget(id) {
    const driver = state.drivers[id];
    driver.showingTarget = !driver.showingTarget;
    
    if (driver.showingTarget) {
        driver.targetPath.addTo(map);
        driver.targetMarker.addTo(map);
        // Ajustar zoom para caber a rota
        const bounds = L.latLngBounds([driver.marker.getLatLng(), driver.targetMarker.getLatLng()]);
        map.fitBounds(bounds, { padding: [50, 50] });
    } else {
        map.removeLayer(driver.targetPath);
        map.removeLayer(driver.targetMarker);
    }
}

function togglePaths() {
    state.showPaths = !state.showPaths;
    Object.values(state.drivers).forEach(d => {
        if (state.showPaths) {
            d.path.addTo(map);
        } else {
            map.removeLayer(d.path);
        }
    });
}

function getRandomColor() {
    const colors = ['#2563eb', '#dc2626', '#059669', '#d97706', '#7c3aed', '#4f46e5'];
    return colors[Math.floor(Math.random() * colors.length)];
}
