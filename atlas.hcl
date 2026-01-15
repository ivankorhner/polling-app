# Atlas migration configuration
# See: https://atlasgo.io/cli/intro

env {
  src = "ent://internal/ent/schema"
  url = getenv("DATABASE_URL")
  dev = "docker://postgres/15/dev?search_path=public"
  
  migration {
    dir = "file://internal/migrate/migrations"
    format = atlas
  }
  
  format {
    migrate {
      diff = "{{ sql . }}"
    }
  }
}
