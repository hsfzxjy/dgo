part of 'generator.dart';

@immutable
class GeneratorSymbol {
  final String _value;
  const GeneratorSymbol(this._value);

  @override
  String toString() => ctx[this];

  @override
  int get hashCode => _value.hashCode;

  @override
  bool operator ==(Object other) =>
      other is GeneratorSymbol && _value == other._value;
}

class GeneratorContext {
  final StringSink buffer;
  final Importer importer;
  final _symbols = <GeneratorSymbol, String>{};
  final _usedNames = <String>{};

  static final _stack = <GeneratorContext>[];

  GeneratorContext(this.buffer, this.importer);

  String operator [](GeneratorSymbol sym) => _symbols[sym]!;

  String pickUnique(String name) {
    var counter = 1;
    var name2 = name;
    while (true) {
      if (!_usedNames.contains(name2)) {
        _usedNames.add(name2);
        return name2;
      }
      name2 = '$name$counter';
      counter++;
    }
  }
}

GeneratorContext get ctx => GeneratorContext._stack.last;
GeneratorContext get currentContext => ctx;

void setFile(File file, Map<GeneratorSymbol, String> symbolMap) {
  assert(GeneratorContext._stack.length <= 1);
  final newContext = GeneratorContext(file, file.importer)
    .._symbols.addAll(symbolMap)
    .._usedNames.addAll(symbolMap.values);
  GeneratorContext._stack
    ..clear()
    ..add(newContext);
}

const vArgs = GeneratorSymbol('#!vArgs');
const vIndex = GeneratorSymbol('#!vIndex');
const vHolder = GeneratorSymbol('#!vHolder');

typedef JsonMap = Map<String, dynamic>;
