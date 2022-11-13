part of 'generator.dart';

final _formatter = DartFormatter(fixes: [StyleFix.singleCascadeStatements]);

class File extends StringBuffer {
  final DartFileUri uri;
  final Importer importer;

  File(this.uri)
      : importer = Importer(uri),
        super() {
    importer.import3Party('package:meta/meta.dart');
    importer.import3Party('package:dgo/dgo.dart');
    importer.import3Party('dart:async');
  }

  void save(String destDir) {
    final destP = p.join(destDir, uri.toString());
    io.Directory(p.dirname(destP)).createSync(recursive: true);
    final buffer = StringBuffer();
    importer.writeTo(buffer);
    buffer.write(this);
    io.File(destP).writeAsStringSync(_formatter.format(buffer.toString()));
  }

  GeneratorContext asGeneratorContext() => GeneratorContext(this, importer);
}
