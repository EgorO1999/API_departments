# API

REST API для работы с отделами.
Реализован на Go с использованием GORM и PostgreSQL. Миграции базы через Goose. 

Методы API:
Подразделения (Departments):
POST /departments — Создать новое подразделение;
GET /departments/{id} — Получить подразделение по ID, включая сотрудников и дочерние подразделения;
PATCH /departments/{id} — Обновить подразделение;
DELETE /departments/{id}?mode=cascade — Каскадное удаление подразделения, удаляются все дочерние подразделения и сотрудники;
DELETE /departments/{id}?mode=reassign&reassign_to_department_id=id — Удаление подразделения с переносом сотрудников в другое подразделение;


Сотрудники (Employees):
POST /departments/{id}/employees — Создать сотрудника в подразделении


Запуск в Docker

В корневой папке выполнить:
docker-compose up --build

После запуска сервер будет доступен по адресу:
http://localhost:8080
