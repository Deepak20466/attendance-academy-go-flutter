class PagedResult<T> {
  PagedResult({required this.data, required this.page, required this.pageSize, required this.totalCount});

  factory PagedResult.fromJson(Map<String, dynamic> json, T Function(Map<String, dynamic>) fromJson) {
    final rawData = (json['data'] as List<dynamic>?) ?? [];
    return PagedResult(
      data: rawData.map((e) => fromJson(e as Map<String, dynamic>)).toList(),
      page: json['page'] as int? ?? 1,
      pageSize: json['page_size'] as int? ?? rawData.length,
      totalCount: json['total_count'] as int? ?? rawData.length,
    );
  }

  final List<T> data;
  final int page;
  final int pageSize;
  final int totalCount;

  bool get hasMore => page * pageSize < totalCount;
}
