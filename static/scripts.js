// static/scripts.js

document.addEventListener("DOMContentLoaded", () => {
    console.log("DOM полностью загружен");

    const startButton = document.getElementById("startButton");
    const accidentButton = document.getElementById("accidentButton");
    const customerCountInput = document.getElementById("customerCount");

    const storeCanvas = document.getElementById("storeCanvas");
    const storeCtx = storeCanvas.getContext("2d");

    const statsCanvas = document.getElementById("statsCanvas");
    const statsCtx = statsCanvas.getContext("2d");

    const goroutineCanvas = document.getElementById("goroutineCanvas");
    const goroutineCtx = goroutineCanvas.getContext("2d");

    const logCanvas = document.getElementById("logCanvas");
    const logCtx = logCanvas.getContext("2d");

    const activeCustomersElem = document.getElementById("activeCustomers");
    const customersByCategoryElem = document.getElementById("customersByCategory");

    const tooltip = document.getElementById("tooltip");

    // Загрузка изображений
    const images = {};
    const imageSources = {
        "Unknown": "/images/unknown.png",
        "Customer": "/images/customer.png",
        "Exit": "/images/exit.png",
        "Aisle": "/images/aisle.png"
    };

    let imagesLoaded = 0;
    const totalImages = Object.keys(imageSources).length;

    for (let key in imageSources) {
        images[key] = new Image();
        images[key].src = imageSources[key];
        images[key].onload = () => {
            imagesLoaded++;
            if (imagesLoaded === totalImages) {
                console.log("Все изображения загружены.");
                drawStoreLayout(); // Отрисовка схемы магазина после загрузки всех изображений
            }
        };
        images[key].onerror = () => {
            console.error(`Не удалось загрузить изображение: ${imageSources[key]}`);
        };
    }

    // Переменная для хранения состояния симуляции
    let simulationState = null;

    // Хранение анимируемых покупателей
    let animatedCustomers = {};

    // Определение отделов магазина
    const departments = [
        { name: "Молочный отдел", position: { x: 200, y: 100 }, width: 100, height: 200 },
        { name: "Отдел овощей", position: { x: 400, y: 100 }, width: 100, height: 200 },
        { name: "Отдел мяса", position: { x: 600, y: 100 }, width: 100, height: 200 },
        { name: "Отдел хлеба", position: { x: 200, y: 350 }, width: 100, height: 200 },
        { name: "Отдел сахара", position: { x: 400, y: 350 }, width: 100, height: 200 }
    ];

    // Установка WebSocket-соединения
    let socket = null;

    function initializeWebSocket() {
        socket = new WebSocket("ws://localhost:8080/ws");

        socket.onopen = () => {
            console.log("WebSocket соединение установлено.");
        };

        socket.onmessage = (event) => {
            console.log("Получено сообщение от сервера:", event.data);
            try {
                const data = JSON.parse(event.data);
                if (data.customers) {
                    simulationState = data;
                    updateStats(simulationState);
                    updateAnimatedCustomers(simulationState.customers);
                }
            } catch (error) {
                console.error("Ошибка парсинга JSON данных:", error);
            }
        };

        socket.onerror = (error) => {
            console.error("Ошибка WebSocket:", error);
        };

        socket.onclose = () => {
            console.log("WebSocket соединение закрыто.");
        };
    }

    initializeWebSocket();

    // Обработчик события для кнопки "Старт"
    startButton.addEventListener("click", () => {
        const customerCount = parseInt(customerCountInput.value);

        console.log("Кнопка 'Старт' нажата.");
        console.log("Количество покупателей:", customerCount);

        if (isNaN(customerCount) || customerCount < 1) {
            console.error("Некорректные значения для запуска симуляции.");
            alert("Пожалуйста, введите корректное значение для количества покупателей.");
            return;
        }

        const message = {
            action: "start",
            data: {
                customerCount: customerCount
            }
        };

        if (socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(message));
            console.log("Сообщение отправлено на сервер:", message);
        } else {
            console.error("WebSocket соединение не установлено.");
            alert("Соединение с сервером не установлено. Попробуйте перезагрузить страницу.");
            return;
        }

        // Очистка Canvas при запуске симуляции
        clearStoreCanvas();
        clearStatsCanvas();
        clearGoroutineCanvas();
        clearLogCanvas();
        animatedCustomers = {}; // Очистка анимированных покупателей

        // Удаление всех иконок покупателей из предыдущих запусков
        const existingIcons = document.querySelectorAll('.customer-icon');
        existingIcons.forEach(icon => icon.remove());

        // Перерисовка схемы магазина
        drawStoreLayout();
    });

    // Обработчик события для кнопки "Авария"
    accidentButton.addEventListener("click", () => {
        if (accidentButton.textContent === "Запустить аварии") {
            fetch('/api/accidents/start', { method: 'POST' })
                .then(response => {
                    if (!response.ok) {
                        throw new Error("Не удалось запустить аварии.");
                    }
                    return response.json();
                })
                .then(data => {
                    console.log(data.message);
                    accidentButton.textContent = "Остановить аварии";
                })
                .catch(error => console.error("Ошибка запуска аварий:", error));
        } else {
            fetch('/api/accidents/stop', { method: 'POST' })
                .then(response => {
                    if (!response.ok) {
                        throw new Error("Не удалось остановить аварии.");
                    }
                    return response.json();
                })
                .then(data => {
                    console.log(data.message);
                    accidentButton.textContent = "Запустить аварии";
                })
                .catch(error => console.error("Ошибка остановки аварий:", error));
        }
    });

    // Функции для очистки Canvas
    function clearStoreCanvas() {
        storeCtx.clearRect(0, 0, storeCanvas.width, storeCanvas.height);
        storeCtx.fillStyle = "#f0f0f0";
        storeCtx.fillRect(0, 0, storeCanvas.width, storeCanvas.height);
        drawStoreLayout(); // Перерисовка схемы магазина
    }

    function clearStatsCanvas() {
        statsCtx.clearRect(0, 0, statsCanvas.width, statsCanvas.height);
        statsCtx.fillStyle = "#ffffff";
        statsCtx.fillRect(0, 0, statsCanvas.width, statsCanvas.height);
    }

    function clearGoroutineCanvas() {
        goroutineCtx.clearRect(0, 0, goroutineCanvas.width, goroutineCanvas.height);
        goroutineCtx.fillStyle = "#ffffff";
        goroutineCtx.fillRect(0, 0, goroutineCanvas.width, goroutineCanvas.height);
    }

    function clearLogCanvas() {
        logCtx.clearRect(0, 0, logCanvas.width, logCanvas.height);
        logCtx.fillStyle = "#ffffff";
        logCtx.fillRect(0, 0, logCanvas.width, logCanvas.height);
    }

    // Функция для отрисовки схемы магазина
    function drawStoreLayout() {
        // Отрисовка отделов
        departments.forEach(dept => {
            // Отрисовка прямоугольника отдела
            storeCtx.fillStyle = "#d3d3d3";
            storeCtx.fillRect(dept.position.x, dept.position.y, dept.width, dept.height);
            storeCtx.strokeStyle = "#000000";
            storeCtx.strokeRect(dept.position.x, dept.position.y, dept.width, dept.height);

            // Добавление названия отдела
            storeCtx.fillStyle = "#000000";
            storeCtx.font = "14px Arial";
            storeCtx.fillText(dept.name, dept.position.x + 10, dept.position.y - 10);
        });

        // Отрисовка выхода (позади Canvas)
        // Выход уже добавлен как изображение в HTML, поэтому здесь ничего не требуется
    }

    // Функция для обновления анимированных покупателей
    function updateAnimatedCustomers(customers) {
        customers.forEach(customer => {
            if (!animatedCustomers[customer.id]) {
                // Создание нового элемента иконки покупателя
                const customerIcon = document.createElement('img');
                customerIcon.src = images["Customer"].src; // Иконка покупателя
                customerIcon.id = `customer-${customer.id}`;
                customerIcon.className = "icon customer-icon";
                customerIcon.style.left = `${customer.current_position.x - 25}px`; // Центрирование иконки (предполагается, что иконка 50x50)
                customerIcon.style.top = `${customer.current_position.y - 25}px`;
                document.getElementById("simulationContainer").appendChild(customerIcon);

                // Инициализация нового покупателя
                animatedCustomers[customer.id] = {
                    id: customer.id,
                    desiredProduct: customer.desired_product.category,
                    currentPosition: { ...customer.current_position },
                    targetPosition: getDepartmentPosition(customer.desired_product.category),
                    image: images["Customer"],
                    animationFrame: null
                };
            }

            // Обновление целевой позиции
            animatedCustomers[customer.id].targetPosition = getDepartmentPosition(customer.desired_product.category);
        });

        // Удаление покупателей, которых нет в текущем состоянии
        for (let id in animatedCustomers) {
            if (!customers.find(c => c.id === parseInt(id))) {
                // Удаление иконки
                const icon = document.getElementById(`customer-${id}`);
                if (icon) {
                    icon.remove();
                }
                // Отмена анимации
                cancelAnimationFrame(animatedCustomers[id].animationFrame);
                delete animatedCustomers[id];
            }
        }

        // Начало анимации для каждого покупателя
        for (let id in animatedCustomers) {
            animateCustomer(animatedCustomers[id]);
        }
    }

    // Функция для получения позиции отдела по категории продукта
    function getDepartmentPosition(category) {
        const department = departments.find(dept => dept.name.includes(category));
        if (department) {
            // Возвращаем центр отдела
            return {
                x: department.position.x + department.width / 2,
                y: department.position.y + department.height / 2
            };
        }
        // Если отдел не найден, направляемся к выходу
        return { x: 750, y: 575 }; // Позиция выхода
    }

    // Функция для анимации движения покупателя
    function animateCustomer(customer) {
        const speed = 2; // Пикселей за кадр

        function step() {
            const dx = customer.targetPosition.x - customer.currentPosition.x;
            const dy = customer.targetPosition.y - customer.currentPosition.y;
            const distance = Math.sqrt(dx * dx + dy * dy);

            if (distance < speed) {
                customer.currentPosition.x = customer.targetPosition.x;
                customer.currentPosition.y = customer.targetPosition.y;

                // После достижения цели, направляемся к выходу
                customer.targetPosition = { x: 750, y: 575 };
            } else {
                customer.currentPosition.x += (dx / distance) * speed;
                customer.currentPosition.y += (dy / distance) * speed;
            }

            // Обновление позиции иконки
            const icon = document.getElementById(`customer-${customer.id}`);
            if (icon) {
                icon.style.left = `${customer.currentPosition.x - 25}px`; // Центрирование иконки
                icon.style.top = `${customer.currentPosition.y - 25}px`;
            }

            // Запрос следующего кадра
            customer.animationFrame = requestAnimationFrame(step);
        }

        // Проверка, есть ли уже анимация
        if (!customer.animationFrame) {
            step();
        }
    }

    // Функция для обновления статистики
    function updateStats(state) {
        activeCustomersElem.textContent = state.customers.length;

        let categoryStats = '';
        for (let category in state.customer_categories) {
            categoryStats += `${category}: ${state.customer_categories[category]} `;
        }
        customersByCategoryElem.textContent = categoryStats.trim() || '-';

        // Обновление статистики на Canvas
        drawStatsCanvas(state);
        drawGoroutineCanvas(state);
        drawLogCanvas(state);
    }

    function drawStatsCanvas(state) {
        clearStatsCanvas();
        statsCtx.fillStyle = "#000000";
        statsCtx.font = "16px Arial";
        statsCtx.fillText(`Среднее количество покупок: ${state.average_purchase_count.toFixed(2)}`, 10, 30);
        statsCtx.fillText(`Нагрузка магазина: ${state.store_load.toFixed(2)}`, 10, 60);
    }

    function drawGoroutineCanvas(state) {
        clearGoroutineCanvas();
        goroutineCtx.fillStyle = "#000000";
        goroutineCtx.font = "16px Arial";
        goroutineCtx.fillText(`Количество горутин: ${state.goroutine_count}`, 10, 30);
        goroutineCtx.fillText(`Количество каналов: ${state.channel_count}`, 10, 60);
    }

    function drawLogCanvas(state) {
        clearLogCanvas();
        logCtx.fillStyle = "#000000";
        logCtx.font = "12px Arial";
        state.technical_logs.slice(-10).forEach((log, index) => {
            logCtx.fillText(log, 10, 20 + index * 15);
        });
    }

    // Обработчик событий для отображения подсказок
    storeCanvas.addEventListener('mousemove', function(event) {
        const rect = storeCanvas.getBoundingClientRect();
        const mouseX = event.clientX - rect.left;
        const mouseY = event.clientY - rect.top;

        let hoveredCustomer = null;

        // Проверка, находится ли курсор над каким-либо покупателем
        if (simulationState && simulationState.customers) {
            simulationState.customers.forEach(customer => {
                const x = customer.current_position.x;
                const y = customer.current_position.y;
                const radius = 25; // Размер иконки

                const distance = Math.sqrt(Math.pow(mouseX - x, 2) + Math.pow(mouseY - y, 2));
                if (distance <= radius) {
                    hoveredCustomer = customer;
                }
            });
        }

        if (hoveredCustomer) {
            tooltip.style.left = (event.clientX + 10) + 'px';
            tooltip.style.top = (event.clientY + 10) + 'px';
            tooltip.innerHTML = `ID: ${hoveredCustomer.id}<br>Отдел: ${hoveredCustomer.desired_product.category}`;
            tooltip.style.display = 'block';
        } else {
            tooltip.style.display = 'none';
        }
    });

    // Скрытие подсказки при уходе курсора с Canvas
    storeCanvas.addEventListener('mouseleave', function() {
        tooltip.style.display = 'none';
    });

    // Инициализация Swagger UI
    const ui = SwaggerUIBundle({
        url: "/swagger/doc.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout"
    });
});