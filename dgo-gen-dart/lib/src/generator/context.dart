part of 'generator.dart';

@immutable
class GeneratorSymbol {
  final String _value;
  const GeneratorSymbol(this._value);

  @override
  String toString() => _value;

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

  GeneratorContext(this.buffer, this.importer);

  String operator [](GeneratorSymbol sym) => _symbols[sym]!;

  GeneratorContext withSymbol(GeneratorSymbol sym, String value) =>
      GeneratorContext(buffer, importer)
        .._symbols.addAll(_symbols)
        .._symbols[sym] = value;

  GeneratorContext withSymbols(Map<GeneratorSymbol, String> syms) =>
      GeneratorContext(buffer, importer)
        .._symbols.addAll(_symbols)
        .._symbols.addAll(syms);
}

const vArgs = GeneratorSymbol('#!vArgs');
const vIndex = GeneratorSymbol('#!vIndex');
const vHolder = GeneratorSymbol('#!vHolder');

typedef JsonMap = Map<String, dynamic>;