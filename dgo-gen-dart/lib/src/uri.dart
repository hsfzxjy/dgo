// Package imports:
import 'package:meta/meta.dart';
import 'package:path/path.dart' as p;

// Project imports:
import 'generator/generator.dart' show config;

@immutable
class GoModUri {
  final String name;
  final String subPath;

  static const GoModUri empty = GoModUri(name: '', subPath: '');

  const GoModUri({
    required this.name,
    required this.subPath,
  });

  String get dartDirPath => p.join(config.renames[name] ?? name, subPath);

  DartFileUri dartFile(String fileName) =>
      DartFileUri(goMod: this, fileName: fileName);
  DartFileUri get dartModFile => dartFile('module.dart');

  @override
  int get hashCode => name.hashCode + subPath.hashCode;

  @override
  bool operator ==(Object other) {
    if (other is GoModUri) {
      return other.name == name && other.subPath == subPath;
    } else {
      return false;
    }
  }
}

abstract class HasGoModUri {
  GoModUri get goMod;
}

@immutable
class DartFileUri implements HasGoModUri {
  @override
  final GoModUri goMod;
  final String fileName;

  const DartFileUri({required this.goMod, required this.fileName});
  const DartFileUri.fromFileName(this.fileName) : goMod = GoModUri.empty;

  @override
  String toString() => p.join(goMod.dartDirPath, fileName);

  String relativeTo(GoModUri from) =>
      p.relative(toString(), from: from.dartDirPath);

  @override
  int get hashCode => goMod.hashCode + fileName.hashCode;

  @override
  bool operator ==(Object other) {
    if (other is DartFileUri) {
      return other.fileName == fileName && other.goMod == goMod;
    } else {
      return false;
    }
  }
}

@immutable
class EntryUri implements HasGoModUri {
  @override
  final GoModUri goMod;
  final String name;

  EntryUri replaceName(String newName) => EntryUri(goMod: goMod, name: newName);

  const EntryUri({
    required this.goMod,
    required this.name,
  });

  factory EntryUri.fromString(String raw) {
    final parts = raw.split('#');
    final goMod = GoModUri(name: parts[0], subPath: parts[1]);
    return EntryUri(goMod: goMod, name: parts[2]);
  }
}
