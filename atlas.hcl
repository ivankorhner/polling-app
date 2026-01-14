# Atlas migration configuration
# See: https://atlasgo.io/cli/intro

env "local" {
  src = "ent://internal/ent/schema"
  dev = "sqlite://?mode=memory"
  
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

env "docker" {
  src = "ent://internal/ent/schema"
  dev = "docker://postgres"
  
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
