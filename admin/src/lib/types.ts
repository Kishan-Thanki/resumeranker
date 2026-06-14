export type AnalysisStatus = 'queued' | 'processing' | 'completed' | 'failed';

export type SectionId = 'skills' | 'experience' | 'education' | 'leadership';

export type MatchStrength = 'strong' | 'partial' | 'weak' | 'none';

export interface Evidence {
	text: string;
	source: 'jd' | 'resume';
	location?: string;
}

export interface RequirementMatch {
	id: string;
	requirement: string;
	jdEvidence: Evidence;
	matched: boolean;
	matchStrength: MatchStrength;
	resumeEvidence: Evidence[];
	note?: string;
}

export interface SectionScore {
	id: SectionId;
	label: string;
	score: number;
	requirements: RequirementMatch[];
}

export interface AnalysisResult {
	id: string;
	createdAt: string;
	jdTitle: string;
	jdText: string;
	resumeName: string;
	status: AnalysisStatus;
	sections: SectionScore[];
	errorMessage?: string;
}
