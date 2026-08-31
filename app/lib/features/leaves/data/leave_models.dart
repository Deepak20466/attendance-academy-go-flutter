class LeaveModel {
  LeaveModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        coachId = json['coach_id'] as String,
        startDate = DateTime.parse(json['start_date'] as String),
        endDate = DateTime.parse(json['end_date'] as String),
        reason = json['reason'] as String,
        status = json['status'] as String;

  final String id;
  final String coachId;
  final DateTime startDate;
  final DateTime endDate;
  final String reason;
  final String status;
}
