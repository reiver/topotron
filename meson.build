project('topotron',
  version: '0.1.0',
  meson_version: '>= 0.64.0',
)

gnome = import('gnome')
i18n = import('i18n')

APP_ID = 'link.reiver.topotron'
APP_PATH = '/link/reiver/topotron'
PKGDATA_DIR = join_paths(get_option('prefix'), get_option('datadir'), meson.project_name())

dependency('glib-2.0')
dependency('gtk4', version: '>= 4.10.0')
dependency('libadwaita-1', version: '>= 1.4')

# Find Go compiler
go_root = run_command('printenv', 'GOROOT', check: false)
if go_root.returncode() != 0
  go_bin = find_program('go')
else
  go_bin = find_program(join_paths(go_root.stdout().strip(), 'bin', 'go'))
endif

subdir('data')
subdir('po')

# Build Go binary
custom_target('topotron',
  build_by_default: true,
  build_always_stale: true,
  output: 'topotron',
  console: true,
  install: true,
  install_dir: get_option('bindir'),
  command: [go_bin, 'build', '-v', '-mod=vendor', '-o', '@OUTPUT@',
            meson.current_source_dir()]
)

gnome.post_install(
  glib_compile_schemas: true,
  gtk_update_icon_cache: true,
  update_desktop_database: true,
)
