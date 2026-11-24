# A test mod for TestPermittedDatabaseRoutes (backend/permitted_routes_test.go): it tries every route a mod
# has to reach a database or a file through a database. The test copies this folder, fills the @OUTSIDE@
# placeholder in its files, makes the copy the permitted folder and gives `database` through the eval
# context. A new route is one row in TestPermittedDatabaseRoutes.
mod "permitted_db_routes" {
  title    = "permitted_db_routes"
  database = var.database
}
