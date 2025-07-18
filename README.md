Сервис для задания L0

Краткое описание содержимого:

Dockerfile и docker-compose.yml для сборки Docker-приложения, в которое включены три контейнера: kafka, БД, go-сервис.

1) cache_data - хранение uid тех заказов, которые кэшируются
2) cmd - там main.go
3) internal :
  3.1) cache - реализация кэширования
  3.2) data_base - для работы с БД
  3.3) handlers - обработка запросов
  3.4) kafka - consumer
  3.5) models - структуры данных для заказа и ответа сервиса
4) kafka/setup.sh - инструкции для запуска kafka
5) postgres/init.sql - инициализация базы данных
6) static/index.html - html страничка
7) test_data - json'ы для тестовой отправки через kafka
