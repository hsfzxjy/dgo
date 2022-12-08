part of 'generator.dart';

final _formatter = DartFormatter(fixes: [StyleFix.singleCascadeStatements]);

class File extends StringBuffer {
  final DartFileUri uri;
  final Importer importer;

  File(this.uri)
      : importer = Importer(uri),
        super() {
    importer.import3Party('package:meta/meta.dart', alias: '\$meta');
    importer.import3Party('package:dgo/dgo.dart', alias: '\$dgo');
    importer.import3Party('dart:async', alias: '\$async');
    importer.import3Party('dart:core', alias: '\$core');
  }

  String get path => p.join(config.generatedInPath, uri.toString());

  Future<void> _ensureWritable() async {
    if (!p.isWithin(config.generatedInPath, path)) {
      throw 'dgo-gen-dart: $path goes out of ${config.generatedInPath}';
    }
    await io.Directory(p.dirname(path)).create(recursive: true);
  }

  Future<void> save() async {
    await _ensureWritable();
    final buffer = StringBuffer();
    buffer.writeln(
        '// ignore_for_file: type=lint, unused_local_variable, unused_import');
    importer.writeTo(buffer);
    buffer.write(this);
    await io.File(path).writeAsString(_formatter.format(buffer.toString()));
  }

  GeneratorContext asGeneratorContext() => GeneratorContext(this, importer);
}
