desktop_file = i18n.merge_file(
  input: 'link.reiver.topotron.desktop.in',
  output: 'link.reiver.topotron.desktop',
  type: 'desktop',
  po_dir: '../po',
  install: true,
  install_dir: join_paths(get_option('datadir'), 'applications')
)

desktop_utils = find_program('desktop-file-validate', required: false)
if desktop_utils.found()
  test('Validate desktop file', desktop_utils, args: [desktop_file])
endif

metainfo_file = i18n.merge_file(
  input: 'link.reiver.topotron.metainfo.xml.in',
  output: 'link.reiver.topotron.metainfo.xml',
  po_dir: '../po',
  install: true,
  install_dir: join_paths(get_option('datadir'), 'metainfo')
)

appstream_cli = find_program('appstreamcli', required: false)
if appstream_cli.found()
  test('Validate appstream file', appstream_cli,
    args: ['validate', '--no-net', '--explain', metainfo_file])
endif

install_data('link.reiver.topotron.gschema.xml',
  install_dir: join_paths(get_option('datadir'), 'glib-2.0/schemas')
)

gnome.compile_resources('topotron',
  'topotron.gresource.xml',
  gresource_bundle: true,
  source_dir: meson.current_build_dir(),
  install: true,
  install_dir: PKGDATA_DIR,
)

# Icons
scalable_dir = join_paths('hicolor', 'scalable', 'apps')
install_data(
  join_paths('icons', scalable_dir, 'link.reiver.topotron.svg'),
  install_dir: join_paths(get_option('datadir'), 'icons', scalable_dir)
)

symbolic_dir = join_paths('hicolor', 'symbolic', 'apps')
install_data(
  join_paths('icons', symbolic_dir, 'link.reiver.topotron-symbolic.svg'),
  install_dir: join_paths(get_option('datadir'), 'icons', symbolic_dir)
)
