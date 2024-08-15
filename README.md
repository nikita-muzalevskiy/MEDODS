# MEDODS. Реализация части сервиса аутентиикации.

# 1. База данных.

Мой запрос на создание таблицы:
CREATE TABLE public."Users"
(
    id bigserial NOT NULL,,
    guid uuid NOT NULL,
    refresh character varying COLLATE pg_catalog."default",
    email character varying COLLATE pg_catalog."default",
    CONSTRAINT "Users_pkey" PRIMARY KEY (id, guid)
);

Мой запрос на добавление записи в таблицу:
insert into "Users" values (default, uuid_generate_v4(), null, null);

Предварительно для использования функции uuid_generate_v4() необходимо выполнить следующую команду:
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

# 2. environment.env
Файл содержит необходимые данные, которые надо прописать вручную.
1) Параметры для подключения к базе данных (пароль, имя пользователя, наименования БД, имя драйвера (у меня было 'postgres'));
2) Ключ для подписи access jwt.

# 3. Docker.
Подготовил Dickerfile, использовал следующие команды в терминале:
1) Сборка
docker build -t medods-server .
2) Запуск
docker run -d --name=medods-server-container -p 80:8080 medods-server
3) Посмотреть статус
docker ps -a

К Сервис работал, но не удавалось подклюиться к БД. Я так предполагаю, что его тоже надо обернуть в докер, а потом изменить параметры подключения к БД. К сожалению, не успел.

# 4. Запросы.
Запросы делал через Postman.
1) Получение токенов по guid:
URL: http://localhost:8080/auth (POST)
Тело запроса:
{
    "UserGuid": "fe3beac6-bacc-4937-94e0-f9b89d95ebc5"
}
id взят в качестве примера.
Если guid нет в БД, то токены не выдадут.

2) Обновление токенов:
URL: http://localhost:8080/refresh (POST)
Тело запроса - ответ, полученный при авторизации. Пример:
{
    "accessToken": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1c2VyR3VpZCI6ImZlM2JlYWM2LWJhY2MtNDkzNy05NGUwLWY5Yjg5ZDk1ZWJjNSIsInVzZXJJcCI6Ils6OjFdOjQ5NDU0In0.kcelhhfiIw3yDjsWIFy438AJCu35D9lQ5L1W4AG54VM7EtEVOD4sRHLkhpZf9bDmqKBIFSs434j3yAqjyG6_hA",
    "refreshToken": "NDk5ZDc4ODEtYjZmNi00MTI5LWJiZTktYmQwMjk2ZTJlYjYz"
}

# 4. IP
При получении запроса на обновление токенов, сравниваются старый и новый IP-адреса. Если они отличаются, в лог выводится сообщения о том, что мы как будто отправили сообщение на email.

P.S. Благодарю за задание, оно было очень полезным и интересным! Надеюсь, моя реализация поможет мне пройти на следующий этап :)



