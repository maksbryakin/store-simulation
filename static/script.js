document.addEventListener("DOMContentLoaded", () => {
    console.log("DOM полностью загружен");

    const startButton = document.getElementById("startButton");
    const customerCountInput = document.getElementById("customerCount");
    const productRangeInput = document.getElementById("productRange");
    const canvas = document.getElementById("storeCanvas");
    const ctx = canvas.getContext("2d");

    if (!startButton) {
        console.error("Кнопка 'Старт' не найдена!");
        return;
    }

    if (!customerCountInput) {
        console.error("Поле 'customerCount' не найдено!");
        return;
    }

    if (!productRangeInput) {
        console.error("Поле 'productRange' не найдено!");
        return;
    }

    if (!canvas) {
        console.error("Элемент Canvas не найден!");
        return;
    }

    if (!ctx) {
        console.error("Не удалось получить контекст Canvas.");
        return;
    }

    // Переменная для хранения состояния симуляции
    let simulationState = null;

    // Установка WebSocket-соединения
    const socket = new WebSocket("ws://localhost:8080/ws");

    socket.onopen = () => {
        console.log("WebSocket соединение установлено.");
    };

    socket.onmessage = (event) => {
        console.log("Получено сообщение от сервера:", event.data);
        const data = JSON.parse(event.data);
        if (data.Customers) {
            simulationState = data;
            // Отрисовка на Canvas
            drawSimulation(simulationState);
        }
    };

    socket.onerror = (error) => {
        console.error("Ошибка WebSocket:", error);
    };

    socket.onclose = () => {
        console.log("WebSocket соединение закрыто.");
    };

    // Обработчик события для кнопки "Старт"
    startButton.addEventListener("click", () => {
        const customerCount = parseInt(customerCountInput.value);
        const productRange = productRangeInput.value.split("-").map(Number);

        console.log("Кнопка 'Старт' нажата.");
        console.log("Количество покупателей:", customerCount);
        console.log("Диапазон продуктов:", productRange);

        if (isNaN(customerCount) || productRange.length !== 2) {
            console.error("Некорректные значения для запуска симуляции.");
            return;
        }

        const message = {
            action: "start",
            data: {
                customerCount: customerCount,
                products: {
                    milk: { min: productRange[0], max: productRange[1] },
                    bread: { min: productRange[0], max: productRange[1] },
                    // Добавь другие категории по необходимости
                }
            }
        };

        if (socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(message));
            console.log("Сообщение отправлено на сервер:", message);
        } else {
            console.error("WebSocket соединение не установлено.");
        }

        // Очистка Canvas при запуске симуляции
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.fillStyle = "#f0f0f0";
        ctx.fillRect(0, 0, canvas.width, canvas.height);
    });

    function drawRoutes(customers) {
        customers.forEach(customer => {
            const aislePosition = { x: 300, y: 100 }; // Позиция Aisle
            const exitPosition = { x: 50, y: 500 };   // Позиция Exit

            ctx.beginPath();
            ctx.moveTo(aislePosition.x, aislePosition.y);
            ctx.lineTo(customer.CurrentPosition.X, customer.CurrentPosition.Y);
            ctx.lineTo(exitPosition.x, exitPosition.y);
            ctx.strokeStyle = 'rgba(0, 0, 0, 0.1)';
            ctx.stroke();
            ctx.closePath();
        });
    }

    // Функция для отрисовки симуляции
    function drawSimulation(state) {
        // Очистка Canvas
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.fillStyle = "#f0f0f0";
        ctx.fillRect(0, 0, canvas.width, canvas.height);

        // Отрисовка клиентов
        state.Customers.forEach(customer => {
            ctx.fillStyle = "#FF0000"; // Красный цвет для клиентов
            ctx.beginPath();
            ctx.arc(customer.X, customer.Y, 10, 0, Math.PI * 2, true); // Радиус 10
            ctx.fill();
        });
    }
});