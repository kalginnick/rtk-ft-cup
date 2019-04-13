# Функциональноть полнотекстового поиска

http://localhost:8080/api/v1/search

- Ввод поискового запроса 'транслитом'.
  http://localhost:8080/api/v1/search?query=pobeda
- Ввод поискового запроса в неверной раскладке клавиатуры.
  http://localhost:8080/api/v1/search?query=yjhvf
- Поиск с учетом возможных словоформ.
  http://localhost:8080/api/v1/search?query=%D0%B8%D1%81%D1%82%D0%BE%D1%80%D0%B8%D0%B8
- Поиск с учетом возможных опечаток.
  http://localhost:8080/api/v1/search?query=%D0%BA%D0%B5%D1%80%D0%BF%D0%B8%D1%87
