part of 'generator.dart';

final _reVariable = RegExp(r'^(?<name>[\$\w]+?)(\$(?<id>\d+))?$');

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

  static final _dupCounter = <GeneratorSymbol, int>{};
  DuplicatedGeneratorSymbol get dup {
    final nextSuffix = _dupCounter[this] ?? 1;
    _dupCounter[this] = nextSuffix + 1;
    final dupSymbol = DuplicatedGeneratorSymbol._(this, nextSuffix);

    {
      final matched = _reVariable.firstMatch(ctx[this])!;
      final basename = matched.namedGroup('name')!;
      int suffix = 1;
      String name;
      while (true) {
        name = '$basename\$$suffix';
        if (!ctx._symbols.containsValue(name)) break;
        suffix++;
      }
      ctx._symbols[dupSymbol] = name;
    }

    return dupSymbol;
  }
}

@immutable
class DuplicatedGeneratorSymbol extends GeneratorSymbol {
  final GeneratorSymbol _origin;
  DuplicatedGeneratorSymbol._(this._origin, int suffix)
      : super('${_origin._value}$suffix');

  @override
  DuplicatedGeneratorSymbol get dup => _origin.dup;
}

class GeneratorContext {
  final StringSink buffer;
  final Importer importer;
  final _symbols = <GeneratorSymbol, String>{};

  static final _stack = <GeneratorContext>[];

  GeneratorContext(this.buffer, this.importer);

  String operator [](GeneratorSymbol sym) => _symbols[sym]!;
}

GeneratorContext get ctx => GeneratorContext._stack.last;
GeneratorContext get currentContext => ctx;

void setFile(File file, Map<GeneratorSymbol, String> symbolMap) {
  assert(GeneratorContext._stack.length <= 1);
  final newContext = GeneratorContext(file, file.importer)
    .._symbols.addAll(symbolMap);
  GeneratorContext._stack
    ..clear()
    ..add(newContext);
}

const vArgs = GeneratorSymbol('#!vArgs');
const vIndex = GeneratorSymbol('#!vIndex');
const vHolder = GeneratorSymbol('#!vHolder');

typedef JsonMap = Map<String, dynamic>;
