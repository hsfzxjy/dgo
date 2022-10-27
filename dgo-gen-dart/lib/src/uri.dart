// Package imports:
import 'package:meta/meta.dart';
import 'package:path/path.dart' as p;

@immutable
class GoModUri {
  final String fullName;
  final String name;
  final String subPath;

  static const GoModUri empty = GoModUri(fullName: '', name: '', subPath: '');

  const GoModUri({
    required this.fullName,
    required this.name,
    required this.subPath,
  });

  String get dartDirPath => p.join(name, subPath);

  DartFileUri dartFile(String fileName) =>
      DartFileUri(goMod: this, fileName: fileName);
  DartFileUri get dartModFile => dartFile('mod.dart');

  @override
  int get hashCode => fullName.hashCode + subPath.hashCode;

  @override
  bool operator ==(Object other) {
    if (other is GoModUri) {
      return other.fullName == fullName && other.subPath == subPath;
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
    final goMod = GoModUri(
      fullName: parts[0],
      name: parts[0].split('/').last.replaceFirst(RegExp(r'^/'), ''),
      subPath: parts[1],
    );
    return EntryUri(goMod: goMod, name: parts[2]);
  }
}
