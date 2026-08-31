class SalaryAckModel {
  SalaryAckModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        coachId = json['coach_id'] as String,
        periodMonth = json['period_month'] as int,
        periodYear = json['period_year'] as int,
        amount = (json['amount'] as num?)?.toDouble(),
        status = json['status'] as String;

  final String id;
  final String coachId;
  final int periodMonth;
  final int periodYear;
  final double? amount;
  final String status;
}
